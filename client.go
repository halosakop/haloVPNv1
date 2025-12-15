package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/songgao/water"
)

const maxUDPSize = 65535

func main() {
	server := flag.String("server", "", "server host:port")
	flag.Parse()

	if *server == "" {
		log.Fatalf("Provide -server host:port")
	}

	cfg := water.Config{DeviceType: water.TUN}
	ifce, err := water.New(cfg)
	if err != nil {
		log.Fatalf("TUN create error: %v", err)
	}

	log.Printf("Opened TUN interface: %s", ifce.Name())

	raddr, err := net.ResolveUDPAddr("udp", *server)
	if err != nil {
		log.Fatalf("resolve error: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatalf("dial error: %v", err)
	}

	log.Printf("Connected to server %s", raddr.String())

	// Send valid IPv4 header probe instead of "hello"
	conn.Write([]byte{0x45, 0, 0, 0})

	// UDP -> TUN
	go func() {
		buf := make([]byte, maxUDPSize)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				continue
			}
			v := buf[0] >> 4
			if v == 4 || v == 6 {
				ifce.Write(buf[:n])
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
				continue
			}
			conn.Write(pkt[:n])
		}
	}()

	// keepalive
	go func() {
		for {
			conn.Write([]byte{0x45, 0, 0, 0})
			time.Sleep(5 * time.Second)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
}
