# Overview 
[![Build Status](https://travis-ci.org/lightstar-dev/openlan-go.svg?branch=master)](https://travis-ci.org/lightstar-dev/openlan-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lightstar-dev/openlan-go)](https://goreportcard.com/report/lightstar-dev/openlan-go)
[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Refer to [danieldin95 openlan-py](https://github.com/danieldin95/openlan-py), and now we change cpe to point, ope to vswitch.
    
# Windows
## Install tap-windows6

Download `resource/tap-windows-9.21.2.exe`, then install it. And run Point in Windows by `point.exe -vs:addr x.x.x.x -vs:auth zzz:wwww`. 
    
## Build in Powershell
### Get source code and it's dependents

    go get -u -v github.com/lightstar-dev/openlan-go
    
### Execute building command.

    PS L:\vswitchnlan-go\olv1> go build -o ./resource/point.exe cpe_windows.go

## Configure Windows TAP Device

Goto `Control Panel\Network and Internet\Network Connections`, and find `Ethernet 2`, then you can configure IPAddress for it to access branch site. 

Or Configure by Powershell.

    netsh interface ipv4 show config "Ethernet 2"
    netsh interface ipv4 set address "Ethernet 2" static 192.168.x.b

# Linux
## Start vSwitch on Linux

    [root@localhost olv1]# cat .passowrd
    zzz:wwww
    xxxx:aaaaa
    [root@localhost olv1]# nohup ./resource/vswitch -vs:addr x.x.x.x -if:addr 192.168.x.a/24 &
    [root@localhost olv1]# cat .vswitchtoken
    m64rxofsqkvlb4cj
    
### Show Points

    [root@localhost olv1]# curl -um64rxofsqkvlb4cj: -XGET http://localhost:10082/

### Show Neightbors

    [root@localhost olv1]# curl -um64rxofsqkvlb4cj: -XGET http://localhost:10082/neighbor

## Start Point on Linux

    [root@localhost olv1]# nohup ./resource/point -vs:addr x.x.x.x -vs:auth zzz:wwww -if:addr 192.168.x.b/24 &
    [root@localhost olv1]# ping 192.168.x.a

    
# Start Point on macOS

    AppledeMBP:openlan-go apple$ make darwin
    AppledeMBP:openlan-go apple$ sudo ./resource/point.dw -if:ethdst b2:bb:ba:c0:8a:4d -if:addr 192.168.10.14

### Configure IP address on `utun2`
    
    AppledeMBP:~ apple$ sudo ifconfig utun2 192.168.10.14 192.168.10.11
    AppledeMBP:~ apple$ ping 192.168.10.11
    
### How to get gateway ethernet address for `utun2`

    AppledeMBP:openlan-go apple$ ./resource/pointctl.dw 
    [point]# 
    [point]# open openlan.net hi:hi@123$
    2019/09/30 16:53:00 INFO PointCmd.Connect
    2019/09/30 16:53:00 INFO TcpClient.Connect openlan.net:10002
    2019/09/30 16:53:01 INFO TcpWroker.TryLogin: {"name":"hi","password":"hi@123$"}
    [point]# 2019/09/30 16:53:01 INFO TcpWroker.onHook.login: okay.
    [point]# arp
    arp <source> <destination>
    [point]# 
    [point]# arp 192.168.10.24 192.168.10.11
    42
    [point]# 2019/09/30 16:53:58 INFO PointCmd.onArp: b2:bb:ba:c0:8a:4d on 192.168.10.11
    [point]# exit
    AppledeMBP:openlan-go apple$ ./resource/point.dw -if:ethdst b2:bb:ba:c0:8a:4d
    

