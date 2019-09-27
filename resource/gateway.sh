#!/bin/bash

yourNet="192.168.4.0/24"

privateGw="192.168.10.20"
publicIp="185.243.43.127"

/usr/bin/ip route add $yourNet via $privateGw

[ -e "/usr/bin/firewall-cmd" ] && {
  /usr/bin/firewall-cmd --zone=public --permanent --direct --passthrough ipv4 -t nat -I POSTROUTING -s $yourNet ! -d $yourNet -j SNAT --to-source $publicIp
}
