#!/bin/bash

# Closing bridge's netfilter.

cat >> /etc/sysctl.conf <<EOF
  net.bridge.bridge-nf-call-ip6tables = 0
  net.bridge.bridge-nf-call-iptables = 0
  net.bridge.bridge-nf-call-arptables = 0
EOF

sysctl -p /etc/sysctl.conf

# iptables -t raw -I PREROUTING -i brol-10002 -j NOTRACK