* [中文版](./README.md)
* [English](./README_EN.md)

# Overview 
[![Build Status](https://travis-ci.org/lightstar-dev/openlan-go.svg?branch=master)](https://travis-ci.org/lightstar-dev/openlan-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lightstar-dev/openlan-go)](https://goreportcard.com/report/lightstar-dev/openlan-go)
[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

OpenLan旨在解决局域网数据报文在广域网的传输问题，并建立基于租户的虚拟以太网络。


                   192.168.1.a/24         192.168.1.b/24              192.168.1.c/24
                         |                      |                           |
                       Point --酒店 Wifi--> vSwitch(南京) <---其他 Wifi--- Point
                                                |
                                             互联网
                                                |
                                           vSwitch(上海) - 192.168.1.d/24
                                                |
                       ------------------------------------------------------
                       ^                        ^                           ^
                       |                        |                           |
                   办公 Wifi               家庭 Wifi                 酒店 Wifi     
                       |                        |                           |
                     Point                    Point                       Point
                 192.168.1.e/24           192.168.1.f/24              192.168.1.g/24
                
 如上图分布在南京的接入点：192.168.1.a、192.168.1.c，通过互联网接入在南京的虚拟交换vSwitch；而分布在上海的接入点：192.168.1.e、192.168.1.f、192.168.1.g，通过互联网接入在上海的虚拟交换；在上海的虚拟交换与南京的虚拟交换之间，通过互联网或者MPLS建立直连链路。

# 接入点（Point）
接入点工作在用户侧，每个接入点通过接入vSwitch可以实现节点间的互联互通。目前接入点已经稳定工作在Windows及Linux系统下，MacOS还存在问题。 

# 虚拟交换（vSwitch）
每个接入虚拟交换的Point就像工作在一个物理的交换机下的主机，多个虚拟交换之间通过Link可以实现Point的跨区域互通。虚拟交换需要安装在Linux的发布系统中，例如：CentOS或者Ubuntu。

## 在Windows系统中
### 首先安装虚拟网卡驱动 tap-windows6

下载资源 `resource/tap-windows-9.21.2.exe`, 然后点击安装它。

### 然后你需要在虚拟网卡上配置地址

打开控制面板`Control Panel\Network and Internet\Network Connections`, 然后找到`Ethernet 2`, 给他配置一个的局域网地址。
或者配置它通过`cmd`.

    netsh interface ipv4 show config "Ethernet 2"
    netsh interface ipv4 set address "Ethernet 2" static 192.168.x.b/24

### 最后配置接入认证

    {
     "vs.addr": "www.openlan.xx",
     "vs.auth": "xx:xx@xx",
     "if.addr": "192.168.x.b/24",
     "vs.tls": true
    }
   
 把它保存在文件`.point.json`中，并与程序`point.windows.x86_64.exe`在同一个目录下。 点击执行`point.windwos.x86_64.exe`。

## 在Linux系统中
### 安装OpenLan并运行vSwitch

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

### 运行Point

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
    

# 从源码编译它

    go get -u -v github.com/lightstar-dev/openlan-go  

## 在Linux系统中

    [root@localhost openlan-go]# make

## 在Windwos系统中
    
    L:\openlan-go> go build -o ./resource/point.windows.x86_64.exe main/point_windows.go
