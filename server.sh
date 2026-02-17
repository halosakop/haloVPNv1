#!/bin/bash

INTERFACE=$(ip -o link show | awk -F': ' '{print $2}' | tail -n 1)

WAN_IF="eth0" # Wan rozhranie

if [ -z "$INTERFACE" ]; then
    echo "Chyba: Nepodarilo sa nájsť žiadne sieťové rozhranie."
    exit 1
fi

# 2. Nastavenie IP adries a tunelu
sudo ip link set "$INTERFACE" up
sudo ip addr add 10.0.0.1/32 dev "$INTERFACE"
sudo ip route add 10.0.0.2/32 dev "$INTERFACE"

# 3. Povolenie IP Forwardingu (okamžité a trvalé)
echo 1 | sudo tee /proc/sys/net/ipv4/ip_forward

# Pridá riadok do sysctl.conf, ak tam ešte nie je
grep -qxF 'net.ipv4.ip_forward=1' /etc/sysctl.conf || echo "net.ipv4.ip_forward=1" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
sudo iptables -t nat -A POSTROUTING -s 10.0.0.0/24 -o "$WAN_IF" -j MASQUERADE
sudo iptables -A FORWARD -i "$INTERFACE" -o "$WAN_IF" -j ACCEPT
sudo iptables -A FORWARD -i "$WAN_IF" -o "$INTERFACE" -m state --state RELATED,ESTABLISHED -j ACCEPT

echo "nastavenie prebehlo úspešne."