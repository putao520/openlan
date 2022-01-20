#!/bin/bash

version="$1"
if [ "$version" == "" ]; then
    echo "$0 <version>"
    exit 1
fi

file="openlan-linux-$version.tar"
if [ ! -e "$file" ]; then
    wget https://github.com/danieldin95/openlan/releases/download/v$version/$file -O $file || {
      rm -rf $file
      exit 1
   }
fi

tar -xf $file
cp -rvf openlan-linux-$version/{etc,usr,var} /
rm -rf openlan-linux-$version

systemctl daemon-reload
sysctl -p /etc/sysctl.d/90-openlan.conf

if [ ! -e "/etc/openlan/switch/switch.json" ]; then
    cp -rvf /etc/openlan/switch/switch.json.example /etc/openlan/switch/switch.json
fi
if [ ! -e "/var/openlan/openvpn/dh.pem" ]; then
    openssl dhparam -out /var/openlan/openvpn/dh.pem 2048
fi
if [ ! -e "/var/openlan/openvpn/ta.key" ]; then
    openvpn --genkey --secret /var/openlan/openvpn/ta.key || {
        echo "please install package for openvpn by yum or apt"
    }
fi
