package main

import (
	"log"
	"net"

	"github.com/songgao/water"
)

const mtu = 1500

func isIPv4(pkt []byte) bool {
	return len(pkt) >= 20 && pkt[0]>>4 == 4
}

func main() {
	cfg := water.Config{
		DeviceType: water.TUN,
	}
	ifce, err := water.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("TUN:", ifce.Name())

	server, _ := net.ResolveUDPAddr("udp", "91.99.203.50:51820")
	conn, err := net.DialUDP("udp", nil, server)
	if err != nil {
		log.Fatal(err)
	}

	// UDP -> TUN
	go func() {
		buf := make([]byte, mtu)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				continue
			}
			if !isIPv4(buf[:n]) {
				continue
			}
			ifce.Write(buf[:n])
		}
	}()

	// TUN -> UDP
	pkt := make([]byte, mtu)
	for {
		n, err := ifce.Read(pkt)
		if err != nil {
			continue
		}
		if !isIPv4(pkt[:n]) {
			continue
		}
		conn.Write(pkt[:n])
	}
}
