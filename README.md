# Overview 
[![Build Status](https://travis-ci.org/lightstar-dev/openlan-go.svg?branch=master)](https://travis-ci.org/lightstar-dev/openlan-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lightstar-dev/openlan-go)](https://goreportcard.com/report/lightstar-dev/openlan-go)
[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Latest implement on [lightstar-dev openlan-go](https://github.com/lightstar-dev/openlan-go).


                  <192.168.1.a/24>        <192.168.1.b/24>                <192.168.1.c/24>
               (Branch)Point ----------> vSwitch(NanJing) <-------------- Point(Branch)
                                                |
                                         vSwitch(ShangHai) <192.168.1.d/24>
                                                |
                       ------------------------------------------------------
                       ^                        ^                           ^
                       |                        |                           |
                   Office Wifi               Home Wifi                 Hotel Wifi     
                       |                        |                           |
                     Point                    Point                       Point
                 <192.168.1.e/24>          <192.168.1.f/24>            <192.168.1.g/24>
                  

# Point
The point is endpoint to access OpenLan vswitch, and all points behind the same vswitch can visit each other like local area network. 

## on Windows
### Firstly, Install tap-windows6

Download `resource/tap-windows-9.21.2.exe`, then install it. 

### And Then Configure Windows TAP Device

Goto `Control Panel\Network and Internet\Network Connections`, and find `Ethernet 2`, then you can configure IPAddress for it to access branch site. 

Or Configure by `cmd`.

    netsh interface ipv4 show config "Ethernet 2"
    netsh interface ipv4 set address "Ethernet 2" static 192.168.x.b/24

### Finally, Configure Access Authentication

    {
     "vs.addr": "www.openlan.xx",
     "vs.auth": "xx:xx@xx",
     "if.addr": "192.168.x.b/24",
     "vs.tls": true
    }
   
   Save to file `.point.json` with same directory of  `point.windows.x86_64.exe`. Click right on `point.windwos.x86_64.exe`, and Run as Administrator.

## on Linux
### Install OpenLan and Start vSwitch on Linux

    [root@localhost openlan-go]# ./install.sh
    [root@localhost openlan-go]# 
    [root@localhost openlan-go]# cat /etc/vswitch.json
    {
      "vs.addr": "0.0.0.0:10002",
      "http.addr": "0.0.0.0:10000",
      "if.addr": "192.168.x.a/24",
      "links": [
        {
          "vs.addr": "aa.openlan.xx",
          "vs.auth": "xx:xx@xx",
          "vs.tls": true
        }
      ],
      "tls.crt": "/var/openlan/ca/crt.pem",
      "tls.key": "/var/openlan/ca/private.key",
      "log.file": "/var/log/vswitch.log"
    }
    [root@localhost openlan-go]# systemctl enable vswitch
    [root@localhost openlan-go]# systemctl start vswitch

### Start Point on Linux

    [root@localhost openlan-go]# cat /etc/point.json
    {
      "vs.addr": "ww.openlan.xx",
      "vs.auth": "xx:xx@xx",
      "if.addr": "192.168.x.c/24",
      "log.file": "/var/log/point.log"
    }
    [root@localhost openlan-go]# systemctl enable point
    [root@localhost openlan-go]# systemctl start point
    [root@localhost openlan-go]# ping 192.168.x.a
    

# Building from Source

    go get -u -v github.com/lightstar-dev/openlan-go  

## on Linux

    [root@localhost openlan-go]# make

## on Windows
    
    L:\openlan-go> go build -o ./resource/point.windows.x86_64.exe main/point_windows.go
