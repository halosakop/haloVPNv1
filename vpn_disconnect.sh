#!/bin/bash

VPN_SERVER_PUBLIC="91.99.203.50"

echo "[+] Removing VPN routes..."
sudo route delete -net 0.0.0.0/1 2>/dev/null || true
sudo route delete -net 128.0.0.0/1 2>/dev/null || true
sudo route delete -host $VPN_SERVER_PUBLIC 2>/dev/null || true

echo "[+] Restoring default route..."
LOCAL_GW=$(netstat -rn | awk '$1=="default" {print $2}')
sudo route add default $LOCAL_GW 2>/dev/null || true

echo ""
echo "======================================================="
echo "VPN DISCONNECTED"
echo "======================================================="
