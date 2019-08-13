iptables -I INPUT -p tcp --dport 10002 -j ACCEPT
iptables -I INPUT -p tcp --dport 10082 -j ACCEPT

iptables -t nat -I PREROUTING -d 117.89.132.47 -p tcp -m tcp --dport 10002 -j DNAT --to-destination 192.168.4.151:10002
iptables -t nat -I PREROUTING -d 117.89.132.47 -p tcp -m tcp --dport 10082 -j DNAT --to-destination 192.168.4.151:10082
