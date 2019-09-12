#!/bin/bash

firewall-cmd --zone=public --add-port=10002/tcp --permanent
firewall-cmd --zone=public --add-port=10082/tcp --permanent
firewall-cmd --reload

#iptables -I INPUT -p tcp --dport 10002 -j ACCEPT
#iptables -I INPUT -p tcp --dport 10082 -j ACCEPT
#iptables -t nat -I PREROUTING -d 117.89.132.47 -p tcp -m tcp --dport 10002 -j DNAT --to-destination 192.168.4.151:10002
#iptables -t nat -I PREROUTING -d 117.89.132.47 -p tcp -m tcp --dport 10082 -j DNAT --to-destination 192.168.4.151:10082

cp -rvf ./resources/point /usr/bin
cp -rvf ./resources/vswitch /usr/bin

cp -rvf ./resources/point.cfg /etc
cp -rvf ./resources/point.service /usr/lib/systemd/system

cp -rvf ./resources/vswitch.cfg /etc
cp -rvf ./resources/vswitch.service /usr/lib/systemd/system

[ -e /etc/vswitch.password ] || {
cat > /etc/vswitch.password << EOF
hi:hi@123$
EOF
}

systemctl enable point
systemctl enable vswitch

#systemctl start point
#systemctl start vswitch
