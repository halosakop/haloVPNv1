#!/bin/bash

INTERFACE=$(ifconfig -l | tr ' ' '\n' | grep utun | tail -n 1)

GATEWAY=$(netstat -rn | grep 'default' | awk '{print $2}' | head -n 1)

if [ -z "$INTERFACE" ]; then
    echo "Chyba: Rozhranie utun nenájdené."
    exit 1
fi

 ifconfig "$INTERFACE" inet 10.0.0.2 10.0.0.1 up
 route add -host 10.0.0.1 -interface "$INTERFACE"
SERVER_IP="91.99.203.50"
 route add -host "$SERVER_IP" "$GATEWAY"

 route add -host 8.8.8.8 "$GATEWAY"
 route add -host 1.1.1.1 "$GATEWAY"

 route add -net 0.0.0.0/1 -interface "$INTERFACE"
 route add -net 128.0.0.0/1 -interface "$INTERFACE"

echo "Klient je nastavený."