# Project HaloVPN
## HaloVPN encrypts the connection between client and the server

Point of using HaloVPN is to hide your public IP addres and to encrypt your internet communication.
HaloVPN uses virtual TUN adapters to configure and read packets. These TUN adapters gets configured using [Water](https://github.com/songgao/water) Library by Songgao.     
For encryption HaloVPN is using AES. HaloVPN currently only support MacOS but it will be increased to all OS in future. For more information visit [halovpn.app](https://halovpn.app)

## Settings
```net.ResolveUDPAddr("udp", "90.90.90.90:51820") ```

The IP address of the server needs to be changed in client.go, and the server port is originaly set to 51820.
```
	server, _ := net.ResolveUDPAddr("udp", "ip address of server with port") 
	conn, err := net.DialUDP("udp", nil, server)                 
	if err != nil {
		log.Fatal(err)
	}
```
Server port can be changed in server.go, and can be set to only listen to certain IP addresses.
```
	addr := net.UDPAddr{ // pocuvanie na IPv4 UDP port 51820
		IP:   net.IPv4zero, // 0.0.0.0 (vsetky IPv4 adresy)
		Port: 51820,
	}
```
For AES encryption to work variable key needs to be set for both server.go and client.go
```
const (
	mtu = 1400
	key = "your generated key"
)

```
HaloVPN can also be used with GUI

For automatic client and server configuration client.sh and server.sh scripts can be used. In client.sh script you need to configure the server IP address.
```
SERVER_IP="ip of your server"
```
In clint.sh you can also change the DNS server that will be used. The default DNS server is set to google and cloudflare.
```
 route add -host 8.8.8.8 "$GATEWAY"
 route add -host 1.1.1.1 "$GATEWAY"
```



