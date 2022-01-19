#!/bin/bash

set -x

modprobe="/usr/sbin/modprobe"
if [ ! -e "$modprobe" ]; then
	modprobe="/sbin/modprobe"
fi

# probe kernel mod.
$modprobe bridge
$modprobe br_netfilter
$modprobe xfrm4_mode_tunnel
$modprobe vxlan

# clean older files.
/usr/bin/find /var/openlan/point -type f -delete
/usr/bin/find /var/openlan/openvpn -name '*.status' -delete
