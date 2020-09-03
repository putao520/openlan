# Overview 
[![Build Status](https://travis-ci.org/danieldin95/openlan-go.svg?branch=master)](https://travis-ci.org/danieldin95/openlan-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/danieldin95/openlan-go)](https://goreportcard.com/report/danieldin95/openlan-go)
[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

The OpenLAN project help you to build a local area network via the Internet.  

## Terminology

* OLSW: OpenLAN Switch
* OLAP: OpenLAN Access Point
* NAT: Network Address translation

## Branch Access

                                        OLSW(Central) - 10.1.2.10/24
                                                ^
                                                |   
                                              Wifi(DNAT)
                                                |
                                                |
                       ----------------------Internet-------------------------
                       ^                        ^                           ^
                       |                        |                           |
                   Branch 1                 Branch 2                     Branch 3    
                       |                        |                           |
                      OLAP                      OLAP                         OLAP
                 10.1.2.11/24              10.1.2.12/24                  10.1.2.13/24

## Multiple Area
                
                   192.168.1.20/24                                 192.168.1.22/24
                         |                                                 |
                        OLAP ---- Wifi ---> OLSW(NanJing) <---- Wifi --- OLAP
                                                |
                                                |
                                             Internet 
                                                |
                                                |
                                           OLSW(ShangHai) - 192.168.1.10/24
                                                |
                       ------------------------------------------------------
                       ^                        ^                           ^
                       |                        |                           |
                   Office Wifi               Home Wifi                 Hotel Wifi     
                       |                        |                           |
                     OLAP                     OLAP                         OLAP
                 192.168.1.11/24           192.168.1.12/24              192.168.1.13/24
                  

# What's OLAP? 
A OLAP is the endpoint to access OLSW, and all OLAPs behind the same switch can visit each other like local area network. 

## on Windows
### Firstly, Install tap-windows6

Download [tap-windows-9](https://github.com/danieldin95/openlan-go/releases/download/tap-windows-9/tap-windows-9.21.2.exe), then install it. 

### Finally, Configure access authentication

   New a file by notepad++

    {
      "tenant": "default",
      "vs.addr": "www.openlan.xx",
      "vs.auth": "hi:123456",
      "if.addr": "192.168.1.11/24",
      "vs.tls": true
    }
   
   Save to file `point.json` with same directory of  `point.windows.x86_64.exe`. Click right on `point.windwos.x86_64.exe`, and Run as Administrator.

## on Linux
### Install OLSW and start it.

    [root@office ~]# wget https://github.com/danieldin95/openlan-go/releases/download/v4.3.14/openlan-vswitch-4.3.14-1.el7.x86_64.rpm
    [root@office ~]# yum install ./openlan-vswitch-4.3.14-1.el7.x86_64.rpm
    [root@office ~]# cat /etc/vswitch/vswitch.json
    {
      "crt.dir": "/var/openlan/ca",
      "log.file": "/var/log/vswitch.log",
      "http.dir": "/var/openlan/public",
      "bridge": [
        {
            "tenant": "default",
            "if.addr": "192.168.1.11/24"
        }
      ]
    }
    
  *Note*
 
    if.addr    Configure address of bridge
    crt.dir    The directory saved cert

  Configure tenant's authentication
  
    [root@office ~]# cat /etc/vswitch/password/default.json
    [
      { "name": "hi", "password": "123456" },
      { "name": "hei", "password": "123456" }
    ]
    
  Enable system service and start
    
    [root@office ~]# systemctl enable vswitch
    [root@office ~]# systemctl start vswitch


### Install OLAP and start it.

    [root@home ~]# wget https://github.com/danieldin95/openlan-go/releases/download/v4.3.14/openlan-point-4.3.14-1.el7.x86_64.rpm
    [root@home ~]# yum install ./openlan-point-4.3.14-1.el7.x86_64.rpm
    [root@home ~]# cat /etc/point/point.json
    {
      "tenant": "default",
      "vs.addr": "www.openlan.xx",
      "vs.auth": "hi:123456",
      "vs.tls": true,
      "if.addr": "192.168.1.21/24",
      "log.file": "/var/log/point.log"
    }
    
  Enable system service and start
    
    [root@home ~]# systemctl enable point
    [root@home ~]# systemctl start point
    
  Testing network by pint
  
    [root@home ~]# ping 192.168.1.11

# Building from source

    [root@localhost ~]# go get -u -v github.com/danieldin95/openlan-go  

## on Linux

    [root@localhost openlan-go]# make

## on Windows
    
    L:\openlan-go> go build -o ./resource/point.windows.x86_64.exe main/point_windows.go
