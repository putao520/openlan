#!/usr/bin/env bash

api=$1
token=$(cat /etc/openlan/switch/token)

/usr/bin/curl -u"${token}": -k -XGET https://localhost:10000/api/"${api}"
