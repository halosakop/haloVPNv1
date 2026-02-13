package main

import (
	"crypto/aes"
	"crypto/cipher"
	"github.com/songgao/water"
	"log"
	"net"
)

const (
	mtu = 1400 //  velkost paketu
	key = "\xb5\x2c\x81\xa0\x1d\x5f\xff\x9a\xac\xbb\xe3\x71\x99\x18\x2b\x71\x97\x5a\x00\x66\xe7\x50\x39\x2c\x65\x97\x25\xab\xd6\xc6\x53\x3b"
)

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
	log.Println("nazov TUN interface:", ifce.Name())

	server, _ := net.ResolveUDPAddr("udp", "91.99.203.50:51820") // preloží IP adresu a port servera
	conn, err := net.DialUDP("udp", nil, server)                 //vytvorenie UDP spojenia na server
	if err != nil {
		log.Fatal(err)
	}

	block, _ := aes.NewCipher([]byte(key)) //sifrovanie pomocou AES
	aesgcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesgcm.NonceSize())

	// UDP -> TUN (dekriptovanie a zapis do TUN)
	go func() { //spustenie gorutiny pre prijimanie UDP paketov
		buf := make([]byte, mtu+100)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				continue
			}
			plainText, err := aesgcm.Open(nil, nonce, buf[:n], nil)
			if err != nil {
				continue
			}
			if !isIPv4(plainText) {
				continue
			}

			ifce.Write(plainText)
		}
	}()

	// TUN -> UDP (kriptovanie a posielanie na server)
	pkt := make([]byte, mtu+100)
	for {
		n, err := ifce.Read(pkt)
		if err != nil {
			continue
		}
		if !isIPv4(pkt[:n]) {
			continue
		}

		cipherText := aesgcm.Seal(nil, nonce, pkt[:n], nil)
		conn.Write(cipherText)
	}
}
