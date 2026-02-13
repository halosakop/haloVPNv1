package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	serverAddr, _ := net.ResolveUDPAddr("udp", "91.99.203.50:5000")

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Example DNS request: A query for "google.com"
	// Prebuilt DNS query packet (not human-readable, just works)
	query := []byte{
		0xab, 0xcd, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x06, 'g', 'o', 'o', 'g', 'l', 'e',
		0x03, 'c', 'o', 'm', 0x00, 0x00, 0x01, 0x00, 0x01,
	}

	_, err = conn.Write(query)
	if err != nil {
		panic(err)
	}

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	resp := make([]byte, 1500)
	n, _, err := conn.ReadFromUDP(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println("Got response from relay:", resp[:n])
}
