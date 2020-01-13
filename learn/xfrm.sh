1. Topo.

     100.141        --        200.130
	   |                        |
192.168.209.141  <=====>   192.168.209.130


2. On 141

# --
ip xfrm state flush

ip xfrm state add src 192.168.209.141 dst 192.168.209.130 proto esp spi 0x00000301 mode tunnel auth md5 0x96358c90783bbfa3d7b196ceabe0536b enc des3_ede 0xf6ddb555acfd9d77b03ea3843f2653255afe8eb5573965df
ip xfrm state add src 192.168.209.130 dst 192.168.209.141 proto esp spi 0x00000302 mode tunnel auth md5 0x99358c90783bbfa3d7b196ceabe0536b enc des3_ede 0xffddb555acfd9d77b03ea3843f2653255afe8eb5573965df

# --
ip xfrm policy flush

ip xfrm policy add src 192.168.100.0/24 dst 192.168.200.0/24 dir out ptype main tmpl src 192.168.209.141 dst 192.168.209.130 proto esp mode tunnel
ip xfrm policy add src 192.168.200.0/24 dst 192.168.100.0/24 dir in ptype main tmpl src 192.168.209.130 dst 192.168.209.141 proto esp mode tunnel
ip xfrm policy add src 192.168.200.0/24 dst 192.168.100.0/24 dir fwd ptype main tmpl src 192.168.209.130 dst 192.168.209.141 proto esp mode tunnel

ip xfrm policy ls

# --

sudo ip netns del net00
sudo ip link del veth-local

sudo ip link add veth-local type veth peer name veth-remote

sudo ip link set veth-local up
sudo ip addr add 192.168.100.1/24 dev veth-local
sudo ip route add 192.168.200.0/24 via 192.168.100.1


sudo ip netns add net00
sudo ip link set veth-remote netns net00

sudo ip netns exec net00 ip link set veth-remote up

sudo ip netns exec net00 ip addr add 192.168.100.141/24 dev veth-remote
sudo ip netns exec net00 ip route add 192.168.200.0/24 via 192.168.100.1



3. On 130

# --
ip xfrm state flush

ip xfrm state add src 192.168.209.141 dst 192.168.209.130 proto esp spi 0x00000301 mode tunnel auth md5 0x96358c90783bbfa3d7b196ceabe0536b enc des3_ede 0xf6ddb555acfd9d77b03ea3843f2653255afe8eb5573965df
ip xfrm state add src 192.168.209.130 dst 192.168.209.141 proto esp spi 0x00000302 mode tunnel auth md5 0x99358c90783bbfa3d7b196ceabe0536b enc des3_ede 0xffddb555acfd9d77b03ea3843f2653255afe8eb5573965df
ip xfrm state get src 192.168.209.141 dst 192.168.209.130 proto esp spi 0x00000301

# --
ip xfrm policy flush

ip xfrm policy add src 192.168.100.0/24 dst 192.168.200.0/24 dir in ptype main tmpl src 192.168.209.141 dst 192.168.209.130 proto esp mode tunnel
ip xfrm policy add src 192.168.100.0/24 dst 192.168.200.0/24 dir fwd ptype main tmpl src 192.168.209.141 dst 192.168.209.130 proto esp mode tunnel
ip xfrm policy add src 192.168.200.0/24 dst 192.168.100.0/24 dir out ptype main tmpl src 192.168.209.130 dst 192.168.209.141 proto esp mode tunnel
ip xfrm policy ls


# --

sudo ip netns del net00
sudo ip link del veth-local

sudo ip link add veth-local type veth peer name veth-remote

sudo ip link set veth-local up
sudo ip addr add 192.168.200.1/24 dev veth-local
sudo ip route add 192.168.100.0/24 via 192.168.200.1


sudo ip netns add net00
sudo ip link set veth-remote netns net00
sudo ip netns exec net00 ip link set veth-remote up

sudo ip netns exec net00 ip addr add 192.168.200.130/24 dev veth-remote
sudo ip netns exec net00 ip route add 192.168.100.0/24 via 192.168.200.1





4. On 141

# host2host

ping 192.168.200.130 -s 1500

# net2net

ip netns exec net00 ping 192.168.200.130 -s 1500



5. On 141

[root@141 ~]# tcpdump -i ens33 -p esp -nne
tcpdump: verbose output suppressed, use -v or -vv for full protocol decode
listening on ens33, link-type EN10MB (Ethernet), capture size 262144 bytes
04:30:39.609139 00:0c:29:b1:77:5f > 00:0c:29:61:70:99, ethertype IPv4 (0x0800), length 1514: 192.168.209.141 > 192.168.209.130: ESP(spi=0x00000301,seq=0x3ff2a), length 1480
04:30:39.609332 00:0c:29:b1:77:5f > 00:0c:29:61:70:99, ethertype IPv4 (0x0800), length 118: 192.168.209.141 > 192.168.209.130: ip-proto-50
04:30:39.610770 00:0c:29:61:70:99 > 00:0c:29:b1:77:5f, ethertype IPv4 (0x0800), length 1514: 192.168.209.130 > 192.168.209.141: ESP(spi=0x00000302,seq=0x6d8db), length 1480
04:30:39.610793 00:0c:29:61:70:99 > 00:0c:29:b1:77:5f, ethertype IPv4 (0x0800), length 118: 192.168.209.130 > 192.168.209.141: ip-proto-50
04:30:40.611232 00:0c:29:b1:77:5f > 00:0c:29:61:70:99, ethertype IPv4 (0x0800), length 1514: 192.168.209.141 > 192.168.209.130: ESP(spi=0x00000301,seq=0x3ff2b), length 1480
04:30:40.611392 00:0c:29:b1:77:5f > 00:0c:29:61:70:99, ethertype IPv4 (0x0800), length 118: 192.168.209.141 > 192.168.209.130: ip-proto-50
04:30:40.612994 00:0c:29:61:70:99 > 00:0c:29:b1:77:5f, ethertype IPv4 (0x0800), length 1514: 192.168.209.130 > 192.168.209.141: ESP(spi=0x00000302,seq=0x6d8dc), length 1480
04:30:40.613023 00:0c:29:61:70:99 > 00:0c:29:b1:77:5f, ethertype IPv4 (0x0800), length 118: 192.168.209.130 > 192.168.209.141: ip-proto-50
