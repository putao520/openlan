# Setup VxLAN Network

## Topology

```
                           192.168.10.2/24      192.168.11.2/24
                                  \                   |
                                   \              --br-v11--
                                --br-v10--           /
                                     \              /
                             VNI[10110/10210]  VNI[10111/10211]  
                                        \         /
                                      OLSW(ShenZhen) 10.10.10.100
                                            |
                                         ESPinUDP  
                                            |
                                            V
                          ----------------Internet--------------
                          ^                                    ^
                          |                                    |
                       ESPinUDP                            ESPinUDP
                          |                                    |
                          |                                    |
        10.10.10.101 OLSW(ShangHai)                       OLSW(NanJing) 10.10.10.102
                      /   |                                    |    \
             VNI[10111] VNI[10110]                       VNI[10210] VNI[10211]
                   /      |                                    |        \ 
                  /    --br-v10--                         --br-v10--     \    
      --br-v11-- |        |                                    |          | --br-v11-- 
                          |                                    |                |
                    192.168.10.1/24                     192.168.10.3/24   192.168.11.3/24

```

# Configure OLSW for ShenZhen