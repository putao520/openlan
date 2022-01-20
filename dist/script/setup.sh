#!/bin/bash

set -x

modprobe="/usr/sbin/modprobe"
[ -e "$modprobe" ] || {
    modprobe="/sbin/modprobe"
}

# probe kernel mod.
$modprobe bridge
$modprobe br_netfilter
$modprobe xfrm4_mode_tunnel
$modprobe vxlan

# clean older files.
find="/usr/bin/find"
$find /var/openlan/point -type f -delete
$find /var/openlan/openvpn -name '*.status' -delete
