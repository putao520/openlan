#!/bin/bash

if [ "$1" == "before" ]; then
	/usr/bin/echo "allowed tcp port before running"
	/usr/sbin/iptables -I INPUT -p tcp --dport 10002 -j ACCEPT
	/usr/sbin/iptables -I INPUT -p tcp --dport 10082 -j ACCEPT 
fi

if [ "$1" == "after" ]; then
	/usr/bin/echo "update bridge after running"
	/usr/sbin/iptables -t raw -I PREROUTING -i brol-10002 -j NOTRACK
fi

if [ "$1" == "exit" ]; then
	/usr/bin/echo "clear iptables on exit"
	/usr/sbin/iptables -D INPUT -p tcp --dport 10002 -j ACCEPT
	/usr/sbin/iptables -D INPUT -p tcp --dport 10082 -j ACCEPT 
	/usr/sbin/iptables -t raw -D PREROUTING -i brol-10002 -j NOTRACK
fi
