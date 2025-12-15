package main

import (
	"log"
	"net"
)

func main() {
	listenAddr, err := net.ResolveUDPAddr("udp", ":5000") // server listens on port 5000
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1500)

	log.Println("Relay server started on UDP port 5000")

	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("read error:", err)
			continue
		}

		// Forward incoming packet to Google DNS
		targetAddr, _ := net.ResolveUDPAddr("udp", "8.8.8.8:53")
		remoteConn, err := net.DialUDP("udp", nil, targetAddr)
		if err != nil {
			log.Println("dial error:", err)
			continue
		}

		_, err = remoteConn.Write(buf[:n])
		if err != nil {
			log.Println("forward error:", err)
			remoteConn.Close()
			continue
		}

		// Read response from Google DNS
		resp := make([]byte, 1500)
		rn, _, err := remoteConn.ReadFromUDP(resp)
		remoteConn.Close()
		if err != nil {
			log.Println("response error:", err)
			continue
		}

		// Send back to original client
		_, err = conn.WriteToUDP(resp[:rn], clientAddr)
		if err != nil {
			log.Println("send back error:", err)
		}
	}
}
