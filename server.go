package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/songgao/water"
)

const maxUDPSizes = 65535

func main() {
	port := flag.Int("port", 51820, "UDP listen port")
	flag.Parse()

	// create TUN
	cfg := water.Config{DeviceType: water.TUN}
	ifce, err := water.New(cfg)
	if err != nil {
		log.Fatalf("create tun: %v", err)
	}
	log.Printf("Opened TUN interface: %s", ifce.Name())
	// NOTE: configure IP for this interface after the program starts (see steps)

	// listen UDP
	addr := net.UDPAddr{Port: *port}
	udp, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("listen udp: %v", err)
	}
	log.Printf("Listening on UDP :%d", *port)

	var clientAddr *net.UDPAddr

	// UDP -> TUN
	go func() {
		buf := make([]byte, maxUDPSizes)
		for {
			n, raddr, err := udp.ReadFromUDP(buf)
			if err != nil {
				log.Printf("udp read error: %v", err)
				continue
			}
			// register/update single client address
			if clientAddr == nil || clientAddr.String() != raddr.String() {
				clientAddr = raddr
				log.Printf("Registered client: %v", clientAddr)
			}
			// write payload into TUN (expected to be raw IP packet)
			_, err = ifce.Write(buf[:n])
			if err != nil {
				log.Printf("tun write error: %v", err)
			}
		}
	}()

	// TUN -> UDP
	go func() {
		pkt := make([]byte, maxUDPSizes)
		for {
			n, err := ifce.Read(pkt)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Printf("tun read error: %v", err)
				continue
			}
			if clientAddr == nil {
				// no client yet: drop packet
				continue
			}
			_, err = udp.WriteToUDP(pkt[:n], clientAddr)
			if err != nil {
				log.Printf("udp write error: %v", err)
			}
		}
	}()

	// keepalive to keep NAT mapping alive (if behind NAT)
	go func() {
		ticker := make(chan struct{})
		// simple tick without importing time (small, explicit)
		// use a small goroutine to sleep + send signal
		go func() {
			for {
				// sleep 15s
				time.Sleep(15 * time.Second)
				ticker <- struct{}{}
			}
		}()
		for range ticker {
			if clientAddr != nil {
				_, _ = udp.WriteToUDP([]byte{0}, clientAddr)
			}
		}
	}()

	// wait for ctrl-c
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
	udp.Close()
	ifce.Close()
}
