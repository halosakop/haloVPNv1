// client.go
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

const maxUDPSize = 65535

func main() {
	server := flag.String("server", "", "server address host:port (required)")
	flag.Parse()

	if *server == "" {
		log.Fatalf("provide -server host:port")
	}

	// create TUN
	cfg := water.Config{DeviceType: water.TUN}
	ifce, err := water.New(cfg)
	if err != nil {
		log.Fatalf("create tun: %v", err)
	}
	log.Printf("Opened TUN interface: %s", ifce.Name())
	// configure IP for this interface after startup (see steps)

	// dial UDP to server
	raddr, err := net.ResolveUDPAddr("udp", *server)
	if err != nil {
		log.Fatalf("resolve server: %v", err)
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatalf("dial udp: %v", err)
	}
	defer conn.Close()
	log.Printf("Connected to server %s", raddr.String())

	// send initial registration / keepalive
	_, _ = conn.Write([]byte("hello"))

	// UDP -> TUN
	go func() {
		buf := make([]byte, maxUDPSize)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Printf("udp read err: %v", err)
				continue
			}
			_, err = ifce.Write(buf[:n])
			if err != nil {
				log.Printf("tun write err: %v", err)
			}
		}
	}()

	// TUN -> UDP
	go func() {
		pkt := make([]byte, maxUDPSize)
		for {
			n, err := ifce.Read(pkt)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Printf("tun read err: %v", err)
				continue
			}
			// send raw IP packet to server
			_, err = conn.Write(pkt[:n])
			if err != nil {
				log.Printf("udp write err: %v", err)
			}
		}
	}()

	// keepalive (simple) - can be replaced with better logic
	go func() {
		for {
			_, _ = conn.Write([]byte{0})
			time.Sleep(15 * time.Second)
		}
	}()

	// wait for ctrl-c
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Println("stopping client")
}
