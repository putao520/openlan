iptables -I INPUT -p tcp --dport 10002 -j ACCEPT
iptables -I INPUT -p tcp --dport 10082 -j ACCEPT

# Disabled bridge-nf-call-iptabless

iptables -t nat -I PREROUTING -d 117.89.132.47 -p tcp -m tcp --dport 10002 -j DNAT --to-destination 192.168.4.151:10002
iptables -t nat -I PREROUTING -d 117.89.132.47 -p tcp -m tcp --dport 10000 -j DNAT --to-destination 192.168.4.151:10000

iptables -t nat -I POSTROUTING -s 192.168.4.0/24 -d 192.168.10.0/24 -j SNAT --to-source 192.168.10.20
iptables -t nat -I POSTROUTING -s 192.168.4.0/24 -d 192.168.10.0/24 -j SNAT --to-source 192.168.10.21
