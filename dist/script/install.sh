#!/bin/bash

set -x

version="$1"
[ "$version" != "" ] || {
    echo "$0 <version>"
    exit 1
}

file="openlan-linux-$version.tar"

function download() {
    [ -e "$file" ] || {
        wget https://github.com/danieldin95/openlan/releases/download/v$version/$file -O $file || {
            rm -rf $file
            exit 1
        }
    }
}

function install() {
   tar -xf $file -C /tmp
   cp -rf  /tmp/openlan-linux-$version/{etc,usr,var} /
   cd /tmp/openlan-linux-$version && find ./ -type f > /usr/share/openlan.db
   rm -rf /tmp/openlan-linux-$version
}

function post() {
  [ -e "/etc/openlan/switch/switch.json" ] || {
      cp -rvf /etc/openlan/switch/switch.json.example /etc/openlan/switch/switch.json
  }
  [  -e "/var/openlan/openvpn/dh.pem" ] || {
      openssl dhparam -out /var/openlan/openvpn/dh.pem 2048
  }
  [ -e "/var/openlan/openvpn/ta.key" ] || {
      openvpn --genkey --secret /var/openlan/openvpn/ta.key
  }
}

function finish() {
    systemctl daemon-reload
    sysctl -p /etc/sysctl.d/90-openlan.conf
}

download
install
post
finish
