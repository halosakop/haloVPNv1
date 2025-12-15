package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/exec" // Added to run system commands
	"os/signal"
	"syscall"
	"time"

	"github.com/songgao/water"
)

const maxUDPSizes = 65535
const serverIP = "10.0.0.1"

func main() {
	port := flag.Int("port", 443, "UDP listen port") // Default changed to 443 to match your usage
	flag.Parse()

	// 1. Create TUN Interface
	cfg := water.Config{DeviceType: water.TUN}
	ifce, err := water.New(cfg)
	if err != nil {
		log.Fatalf("create tun: %v", err)
	}
	log.Printf("Opened TUN interface: %s", ifce.Name())

	// 2. CONFIGURE IP (The Fix)
	// We run the system commands here, keeping the interface alive.
	log.Printf("Configuring interface %s with IP %s...", ifce.Name(), serverIP)

	// Command: ip addr add 10.0.0.1/24 dev tun0
	cmdAddr := exec.Command("ip", "addr", "add", serverIP+"/24", "dev", ifce.Name())
	if out, err := cmdAddr.CombinedOutput(); err != nil {
		log.Printf("Error adding IP: %v, output: %s", err, out)
	}

	// Command: ip link set tun0 up
	cmdUp := exec.Command("ip", "link", "set", ifce.Name(), "up")
	if out, err := cmdUp.CombinedOutput(); err != nil {
		log.Printf("Error setting link up: %v, output: %s", err, out)
	}

	// 3. Start UDP Listener
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
			// Update client address dynamically
			if clientAddr == nil || clientAddr.String() != raddr.String() {
				clientAddr = raddr
				log.Printf("New client connected: %v", clientAddr)
			}

			if n > 0 {
				_, err = ifce.Write(buf[:n])
				if err != nil {
					log.Printf("tun write error: %v", err)
				}
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
				continue
			}
			_, err = udp.WriteToUDP(pkt[:n], clientAddr)
			if err != nil {
				log.Printf("udp write error: %v", err)
			}
		}
	}()

	// Keep NAT alive
	go func() {
		for {
			time.Sleep(15 * time.Second)
			if clientAddr != nil {
				udp.WriteToUDP([]byte{0}, clientAddr)
			}
		}
	}()

	// Wait for exit signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Println("Shutting down...")
}
