# Setup VxLAN Network

## Topology

```
                           192.168.10.10/24        192.168.11.10/24
                                    \                     /
                                     \               --br-v11--
                                  --br-v10--            /
                                       \               /
                             VNI[101110/101210]   VNI[101111/101211]  
                                          \         /
                                        OLSW(ShenZhen) 10.10.10.10
                                              |
                                              |
                                           ESPinUDP  
                                              |
                                              V
                          +-----------------Internet---------------+
                          ^                                        ^
                          |                                        |
                       ESPinUDP                                ESPinUDP
                          |                                        |
                          |                                        |
        10.10.10.11 OLSW(ShangHai)                         OLSW(NanJing) 10.10.10.12
                     /       |                               |        \
           VNI[101111]  VNI[101110]                   VNI[101210]   VNI[101211]
                  /          |                               |           \ 
                 /       --br-v10--                      --br-v10--       \    
          --br-v11--         |                               |         --br-v11--+ 
                             |                               |                   |
                     192.168.10.11/24                  192.168.10.12/24    192.168.11.12/24

```

# Configure OLSW for ShenZhen

配置IPSec/ESP网络：

```
[root@olsw-sz ~]# cd /etc/openlan/switch/network
[root@olsw-sz ~]# cat > esp.json <<EOF
{
    "name": "esp",
    "provider": "esp",
    "interface": {
        "address": "10.10.10.10",
        "state": {
          "auth": "5867f96f31ec",
          "crypt": "a0e5535e4121"
        },
        "members": [
            {
                "spi": 1011,
                "peer": "10.10.10.11",
                "state": {
                    "remote": "sh.esp.net"
                }
            },
            {
                "spi": 1012,
                "peer": "10.10.10.12",
                "state": {
                    "remote": "nj.esp.net"
                }
            }
        ]
    }
}
EOF
[root@olsw-sz ~]# openlan cfg co
```

配置VxLAN网络：

```
[root@olsw-sz ~]# cat > vxlan.json <<EOF
{
    "name": "vxlan",
    "provider": "vxlan",
    "interface": {
        "members": [
            {
                "vni": 101110,
                "remote": "10.10.10.11",
                "bridge": "br-v10"
            },
            {
                "vni": 101210,
                "remote": "10.10.10.12",
                "bridge": "br-v10"
            },
            {
                "vni": 101111,
                "remote": "10.10.10.11",
                "bridge": "br-v11"
            },
            {
                "vni": 101211,
                "remote": "10.10.10.12",
                "bridge": "br-v11"
            }
        ]
    }
}
EOF
[root@olsw-sz ~]# 
[root@olsw-sz ~]# openlan cfg co
[root@olsw-sz ~]# systemctl restart openlan-switch
```

# Configure OLSW for ShangHai

配置IPSec/ESP网络：

```
[root@olsw-sh ~]# cd /etc/openlan/switch/network
[root@olsw-sh ~]# cat > esp.json <<EOF
{
    "name": "esp",
    "provider": "esp",
    "interface": {
        "address": "10.10.10.11",
        "state": {
          "auth": "5867f96f31ec",
          "crypt": "a0e5535e4121"
        },
        "members": [
            {
                "spi": 1011,
                "peer": "10.10.10.10",
                "state": {
                    "remote": "sz.esp.net"
                }
            }
        ]
    }
}
EOF
[root@olsw-sh ~]# openlan cfg co
```

配置VxLAN网络：

```
[root@olsw-sh ~]# cat > vxlan.json <<EOF
{
    "name": "vxlan",
    "provider": "vxlan",
    "interface": {
        "members": [
            {
                "vni": 101110,
                "remote": "10.10.10.10",
                "bridge": "br-v10"
            },
            {
                "vni": 101111,
                "remote": "10.10.10.10",
                "bridge": "br-v11"
            }
        ]
    }
}
EOF
[root@olsw-sh ~]# 
[root@olsw-sh ~]# openlan cfg co
[root@olsw-sh ~]# systemctl restart openlan-switch
[root@olsw-sh ~]# ping 10.10.10.10 -c 3
[root@olsw-sh ~]# ping 192.168.10.10 -c 3
[root@olsw-sh ~]# ping 192.168.11.10 -c 3
```

# Configure OLSW for NanJing

配置IPSec/ESP网络：

```
[root@olsw-nj ~]# cd /etc/openlan/switch/network
[root@olsw-nj ~]# cat > esp.json <<EOF
{
    "name": "esp",
    "provider": "esp",
    "interface": {
        "address": "10.10.10.12",
        "state": {
          "auth": "5867f96f31ec",
          "crypt": "a0e5535e4121"
        },
        "members": [
            {
                "spi": 1012,
                "peer": "10.10.10.10",
                "state": {
                    "remote": "sz.esp.net"
                }
            }
        ]
    }
}
EOF
[root@olsw-nj ~]# openlan cfg co
```

配置VxLAN网络：

```
[root@olsw-nj ~]# cat > vxlan.json <<EOF
{
    "name": "vxlan",
    "provider": "vxlan",
    "interface": {
        "members": [
            {
                "vni": 101210,
                "remote": "10.10.10.10",
                "bridge": "br-v10"
            },
            {
                "vni": 101211,
                "remote": "10.10.10.10",
                "bridge": "br-v11"
            }
        ]
    }
}
EOF
[root@olsw-nj ~]# 
[root@olsw-nj ~]# openlan cfg co
[root@olsw-nj ~]# systemctl restart openlan-switch
[root@olsw-nj ~]# ping 10.10.10.10 -c 3
[root@olsw-nj ~]# ping 10.10.10.11 -c 3
[root@olsw-nj ~]# ping 192.168.10.10 -c 3
[root@olsw-nj ~]# ping 192.168.10.11 -c 3
[root@olsw-nj ~]# ping 192.168.11.10 -c 3
[root@olsw-nj ~]# ping 192.168.11.11 -c 3
```

# 
