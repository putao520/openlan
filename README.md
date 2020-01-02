# 语言
* [中文版](./README.md)
* [English](./README_EN.md)

# 概述 
[![Build Status](https://travis-ci.org/danieldin95/openlan-go.svg?branch=master)](https://travis-ci.org/danieldin95/openlan-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/danieldin95/openlan-go)](https://goreportcard.com/report/lightstar-dev/openlan-go)
[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

OpenLan旨在解决局域网数据报文在广域网的传输问题，并建立基于租户的虚拟以太网络。

Case1：

                                       vSwitch(企业中心) - 10.16.1.10/24
                                                ^
                                                |
                                             Wifi(DNAT)
                                                |
                       ------------------------------------------------------
                       ^                        ^                           ^
                       |                        |                           |
                     分支1                    分支2                        分支3     
                       |                        |                           |
                     Point                    Point                       Point
                 10.16.1.11/24             10.16.1.12/24                10.16.1.13/24
                 

Case2：

                   192.168.1.20/24                                 192.168.1.21/24
                         |                                                 |
                       Point --酒店 Wifi--> vSwitch(南京) <---其他 Wifi--- Point
                                                |
                                             互联网
                                                |
                                           vSwitch(上海) - 192.168.1.10/24
                                                |
                       ------------------------------------------------------
                       ^                        ^                           ^
                       |                        |                           |
                   办公 Wifi               家庭 Wifi                 酒店 Wifi     
                       |                        |                           |
                     Point                    Point                       Point
                192.168.1.11/24           192.168.1.12/24             192.168.1.13/24

 
# 接入点（Point）
接入点工作在用户侧，每个接入点通过接入vSwitch可以实现节点间的互联互通。目前接入点已经稳定工作在Windows及Linux系统下，MacOS还存在问题。 

# 虚拟交换（vSwitch）
每个接入虚拟交换的Point就像工作在一个物理的交换机下的主机，多个虚拟交换之间通过Link可以实现Point的跨区域互通。虚拟交换需要安装在Linux的发布系统中，例如：CentOS或者Ubuntu。

## 在Windows系统中
### 首先安装虚拟网卡驱动 tap-windows6

下载资源 `resource/tap-windows-9.21.2.exe`, 然后点击安装它。

### 最后配置接入认证

    {
     "vs.addr": "www.openlan.xx",
     "vs.auth": "xx:xx@xx",
     "if.addr": "192.168.1.11/24",
     "vs.tls": true
    }
   
 把它保存在文件`point.json`中，并与程序`point.windows.x86_64.exe`在同一个目录下。 点击执行`point.windwos.x86_64.exe`。

 *说明*
 
      vs.addr    虚拟交换的地址或者域名
      vs.auth    接入虚拟交换的认证信息，如：password:user@domain
      if.addr    配置本地虚拟网卡地址
      vs.tls     是否启用TLS加密信道


## 在Linux系统中
### 安装OpenLan并运行vSwitch

    [root@localhost openlan-go]# ./install.sh
    [root@localhost openlan-go]# 
    [root@localhost openlan-go]# cat /etc/vswitch/vswitch.json
    {
      "if.addr": "192.168.1.10/24",
      "links": [
        {
          "vs.addr": "aa.openlan.xx",
          "vs.auth": "xx:xx@xx",
          "vs.tls": true
        }
      ],
      "crt.dir": "/var/openlan/ca",
      "log.file": "/var/log/vswitch.log",
      "http.dir": "/var/openlan/public"
    }
    [root@localhost openlan-go]# systemctl enable vswitch
    [root@localhost openlan-go]# systemctl start vswitch

 *说明*
 
      if.addr    配置本地网桥的地址
      links      配置虚拟交换与其他虚拟交换之间链路
      crt.dir    存放信道加密证书的目录
      log.file   配置日志输出文件

### 运行Point

    [root@localhost openlan-go]# cat /etc/point.json
    {
      "vs.addr": "www.openlan.xx",
      "vs.auth": "xx:xx@xx",
      "if.addr": "192.168.1.21/24",
      "log.file": "/var/log/point.log"
    }
    [root@localhost openlan-go]# systemctl enable point
    [root@localhost openlan-go]# systemctl start point
    [root@localhost openlan-go]# ping 192.168.1.11
    

# 从源码编译它

    go get -u -v github.com/danieldin95/openlan-go  

## 在Linux系统中

    [root@localhost openlan-go]# make

## 在Windwos系统中
    
    L:\openlan-go> go build -o ./resource/point.windows.x86_64.exe main/point_windows.go
