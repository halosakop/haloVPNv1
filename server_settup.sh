#!/bin/bash
set -e

echo "[+] Detecting VPN TUN interface..."
TUN_IF=$(ip link | awk -F: '/tun[0-9]/{print $2}' | tr -d ' ' | head -n1)
echo "    Detected: $TUN_IF"

VPN_NET="10.0.0.0/24"
VPN_SERVER_IP="10.0.0.1"
WAN_IFACE=$(ip route | awk '/default/ {print $5}')
PUBLIC_IP=$(ip route get 1.1.1.1 | awk '{print $7; exit}')

echo "[+] Detected WAN interface: $WAN_IFACE"
echo "[+] Detected public IP: $PUBLIC_IP"

echo "[+] Configuring $TUN_IF with $VPN_SERVER_IP..."
sudo ip addr flush dev $TUN_IF
sudo ip addr add $VPN_SERVER_IP/24 dev $TUN_IF
sudo ip link set $TUN_IF up

echo "[+] Enabling IPv4 forwarding..."
sudo sysctl -w net.ipv4.ip_forward=1
echo "net.ipv4.ip_forward=1" | sudo tee /etc/sysctl.conf

echo "[+] Applying SNAT..."
sudo iptables -t nat -F POSTROUTING
sudo iptables -t nat -A POSTROUTING -s $VPN_NET -o $WAN_IFACE -j SNAT --to-source $PUBLIC_IP

echo "[+] Allowing forwarding..."
sudo iptables -A FORWARD -i $TUN_IF -o $WAN_IFACE -j ACCEPT
sudo iptables -A FORWARD -i $WAN_IFACE -o $TUN_IF -j ACCEPT

echo "[+] Saving firewall rules..."
sudo apt-get install -y iptables-persistent
sudo netfilter-persistent save

echo ""
echo "======================================================="
echo "SERVER SETUP COMPLETE"
echo "Now run your VPN server:"
echo "   sudo go run server.go -port 443"
echo "======================================================="
