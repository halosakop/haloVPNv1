#!/bin/bash

VPN_SERVER_PUBLIC="91.99.203.50"
VPN_CLIENT_IP="10.0.0.2"
VPN_SERVER_IP="10.0.0.1"

LOCAL_GW=$(netstat -rn | awk '$1=="default" {print $2}')
echo "[+] Local gateway detected: $LOCAL_GW"

echo "[+] Starting Go VPN client..."
sudo go run client.go -server ${VPN_SERVER_PUBLIC}:443 &
sleep 2

echo "[+] Waiting for utun interface to appear..."
while true; do
    IFACE=$(ifconfig | grep -o "utun[0-9]" | tail -n1)
    if [ ! -z "$IFACE" ]; then
        echo "    Detected: $IFACE"
        break
    fi
    sleep 1
done

echo "[+] Configuring TUN interface..."
sudo ifconfig utun7 inet $VPN_CLIENT_IP $VPN_SERVER_IP up

echo "[+] Adding route exception for VPN server..."
sudo route add -host $VPN_SERVER_PUBLIC $LOCAL_GW 2>/dev/null || true

echo "[+] Adding full tunnel routing..."
sudo route add -net 0.0.0.0/1 $VPN_SERVER_IP
sudo route add -net 128.0.0.0/1 $VPN_SERVER_IP

echo ""
echo "======================================================="
echo "VPN CONNECTED"
echo "Check your IP: curl ifconfig.me"
echo "======================================================="
