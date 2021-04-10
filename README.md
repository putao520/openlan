
简体中文 | [English](./README_EN.md)

# 概述 
[![Build Status](https://travis-ci.org/danieldin95/openlan-go.svg?branch=master)](https://travis-ci.org/danieldin95/openlan-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/danieldin95/openlan-go)](https://goreportcard.com/report/lightstar-dev/openlan-go)
[![GPL 3.0 License](https://img.shields.io/badge/License-GPL%203.0-blue.svg)](LICENSE)

OpenLAN提供一种局域网数据报文在广域网的传输实现，并能够建立多个用户空间的虚拟以太网络。

## 缩略语

* OLSW: OpenLAN Switch，开放局域网交换机
* OLAP: OpenLAN Access Point，开放局域网接入点
* NAT: Network Address Translation, 网络地址转换
* VxLAN: Virtual eXtensible Local Area Network，虚拟扩展局域网
* IPSec/ESP: Encapsulating Security Payload, IPSec安全封装负载

## 功能清单

* 支持多个网络空间划分，为不同的业务提供逻辑网络隔离；
* 支持OLAP或者OpenVPN接入，提供网桥把局域网共享出去；
* 支持IPSec/ESP隧道，以及基于VxLAN的租户网络划分；
* 支持基于用户名密码的接入认证，使用预共享密约对数据报文进行加密；
* 支持TCP/TLS，UDP/KCP，WS/WSS等多种传输协议实现，TCP模式具有较高的性能；
* 支持HTTP/HTTPS，以及SOCKS5等HTTP的正向代理技术，灵活配置代理进行网络穿透；
* 支持基于TCP的端口转发，为防火墙下的主机提供TCP端口代理。


## 分支接入

                                       OLSW(企业中心) - 10.16.1.10/24
                                                ^
                                                |
                                             Wifi(DNAT)
                                                |
                                                |
                       ----------------------Internet-------------------------
                       ^                        ^                           ^
                       |                        |                           |
                     分支1                    分支2                        分支3     
                       |                        |                           |
                     OLAP                     OLAP                         OLAP
                 10.16.1.11/24             10.16.1.12/24                10.16.1.13/24
                 

## 区域互联

                   192.168.1.20/24                                 192.168.1.21/24
                         |                                                 |
                       OLAP -- 酒店 Wifi --> OLSW(南京) <--- 其他 Wifi --- OLAP
                                                |
                                                |
                                             互联网
                                                |
                                                |
                                             OLSW(上海) - 192.168.1.10/24
                                                |
                                                |
                       ------------------------------------------------------
                       ^                        ^                           ^
                       |                        |                           |
                   办公 Wifi               家庭 Wifi                 酒店 Wifi     
                       |                        |                           |
                     OLAP                     OLAP                         OLAP
                192.168.1.11/24           192.168.1.12/24             192.168.1.13/24
