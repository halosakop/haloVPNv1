#!/bin/bash

INTERFACE=$(ifconfig -l | tr ' ' '\n' | grep utun | tail -n 1)

GATEWAY=$(netstat -rn | grep 'default' | awk '{print $2}' | head -n 1)

if [ -z "$INTERFACE" ]; then
    echo "Chyba: Rozhranie utun nenájdené."
    exit 1
fi

sudo ifconfig "$INTERFACE" inet 10.0.0.2 10.0.0.1 up
sudo route add -host 10.0.0.1 -interface "$INTERFACE"
SERVER_IP="91.99.203.50"
sudo route add -host "$SERVER_IP" "$GATEWAY"

sudo route add -host 8.8.8.8 "$GATEWAY"
sudo route add -host 1.1.1.1 "$GATEWAY"

sudo route add -net 0.0.0.0/1 -interface "$INTERFACE"
sudo route add -net 128.0.0.0/1 -interface "$INTERFACE"

echo "Klient je nastavený."