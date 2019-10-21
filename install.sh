#!/bin/bash

systemctl status firewalld && {
  firewall-cmd --permanent --zone=public --add-port=10000/tcp --permanent
  firewall-cmd --permanent --zone=public --add-port=10002/tcp --permanent
  firewall-cmd --reload
}

[ -e '/usr/sbin/iptables' ] && { 
  echo ""
  #iptables -I INPUT -p tcp --dport 10002 -j ACCEPT
  #iptables -I INPUT -p tcp --dport 10082 -j ACCEPT

  #iptables -t nat -I PREROUTING -d 117.89.132.47 -p tcp -m tcp --dport 10002 -j DNAT --to-destination 192.168.4.151:10002
  #iptables -t nat -I PREROUTING -d 117.89.132.47 -p tcp -m tcp --dport 10082 -j DNAT --to-destination 192.168.4.151:10082
}

cp -rvf ./resource/point.linux.x86_64 /usr/bin/point
cp -rvf ./resource/vswitch.linux.x86_64 /usr/bin/vswitch

[ -e /etc/point.cfg ] || cp -rvf ./resource/point.cfg /etc
[ -e /etc/point.json ] || cp -rvf ./resource/point.json /etc
cp -rvf ./resource/point.service /usr/lib/systemd/system

[ -e /etc/vswitch.json ] || cp -rvf ./resource/vswitch.json /etc
[ -e /etc/vswitch.cfg ] || cp -rvf ./resource/vswitch.cfg /etc
cp -rvf ./resource/vswitch.service /usr/lib/systemd/system


mkdir -p /var/openlan && cp -rvf ./public /var/openlan

[ -e /etc/vswitch.password ] || {
cat > /etc/vswitch.password << EOF
hi:hi@123$
EOF
}

#systemctl enable point
#systemctl enable vswitch

#systemctl start point
#systemctl start vswitch
