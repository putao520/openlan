#!/bin/bash

set -ex

# [root@centos ~]# crontab -l
# 0,5,10,15,20,25,30,35,40,45,50,55 * * * * /var/openlan/script/update.sh
# [root@centos ~]#

export OL_VERSION=v6

# /usr/bin/openlan name add --name your.ddns.name
