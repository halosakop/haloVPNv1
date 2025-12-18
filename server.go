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

	addr := net.UDPAddr{Port: 51820}
	udp, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}

	var client *net.UDPAddr

	// UDP -> TUN
	go func() {
		buf := make([]byte, mtu)
		for {
			n, raddr, err := udp.ReadFromUDP(buf)
			if err != nil {
				continue
			}

			if client == nil {
				client = raddr
				log.Println("Registered client:", client)
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
		if err != nil || client == nil {
			continue
		}
		udp.WriteToUDP(pkt[:n], client)
	}
}
