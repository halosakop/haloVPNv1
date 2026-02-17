package main

import (
	"crypto/aes"
	"crypto/cipher"
	"github.com/songgao/water"
	"log"
	"net"
	"os/exec"
)

const (
	mtu = 1400
	key = "kľúč"
)

func main() {
	cfg := water.Config{DeviceType: water.TUN} // vytvorenie TUN interface
	ifce, err := water.New(cfg)
	if err != nil {
		log.Fatal("Nepodarilo sa vytvorit TUN adapter:", err)
	}
	log.Println("TUN interface:", ifce.Name())

	addr := net.UDPAddr{ // pocuvanie na IPv4 UDP port 51820
		IP:   net.IPv4zero, // 0.0.0.0 (vsetky IPv4 adresy)
		Port: 51820,
	}
	udp, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		log.Fatal("Neprebehlo pocuvanie na UDP:", err)
	}
	log.Println("Pocuvanie na UDP:", addr.Port)
	log.Println(ifce.Name())

	block, _ := aes.NewCipher([]byte(key)) //kryptovanie pomocou AES, vytvorenie cipher bloku s klucom
	aesgcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesgcm.NonceSize())

	log.Println(ifce.Name())
	_ = exec.Command("bash", "./server.sh")

	var client *net.UDPAddr

	// UDP -> TUN (dekriptovanie a zapis do TUN)
	go func() {
		buf := make([]byte, mtu+100)
		for {
			n, raddr, err := udp.ReadFromUDP(buf)
			if err != nil {
				log.Println("UDP chyba:", err)
				continue
			}

			// Registrovanie prveho klienta, ktory posle paket
			if client == nil {
				client = raddr
				log.Println("Registrovany client:", client)
			}

			// prepossielanie paketov do TUN interface
			plainText, err := aesgcm.Open(nil, nonce, buf[:n], nil)
			if err != nil {
				continue
			}

			ifce.Write(plainText)
		}
	}()

	// TUN -> UDP (kriptovanie a posielanie na klienta)
	pkt := make([]byte, mtu+100)
	for {
		n, err := ifce.Read(pkt)
		if err != nil {
			log.Println("TUN chyba:", err)
			continue
		}
		if client == nil {
			// zahodi paket, kym sa nezaregistruje klient
			continue
		}

		cipherText := aesgcm.Seal(nil, nonce, pkt[:n], nil)
		udp.WriteToUDP(cipherText, client)
	}
}
