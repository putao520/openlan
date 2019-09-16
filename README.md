# Overview 

Refer to https://github.com/danieldin95/openlan-py , and now we change cpe to point, ope to vswitch.

# Install tap-windows6

Download `resource/tap-windows-9.21.2.exe`, then install it. And run Point in Windows by `point.exe -vs:addr x.x.x.x -vs:auth zzz:wwww`. 

# Build in Powershell

Download dependent sources

    PS L:\vswitchnlan-go\olv1> go get github.com/songgao/water
    PS L:\vswitchnlan-go\olv1> go get github.com/milosgajdos83/tenus
    PS L:\vswitchnlan-go\olv1> go get golang.org/x/sys

Execute building command.

    PS L:\vswitchnlan-go\olv1> go build -o ./resource/point.exe cpe_windows.go

# Configure Windows TAP Device

Goto `Control Panel\Network and Internet\Network Connections`, and find `Ethernet 2`, then you can configure IPAddress for it to access branch site. 

Or Configure by Powershell.

    netsh interface ipv4 show config "Ethernet 2"
    netsh interface ipv4 set address "Ethernet 2" static 192.168.x.b

# Start vSwitch on Linux

    [root@localhost olv1]# cat .passowrd
    zzz:wwww
    xxxx:aaaaa
    [root@localhost olv1]# nohup ./resource/vswitch -vs:addr x.x.x.x -if:addr 192.168.x.a/24 &
    [root@localhost olv1]# cat .vswitchtoken
    m64rxofsqkvlb4cj
    
Show Points

    [root@localhost olv1]# curl -um64rxofsqkvlb4cj: -XGET http://localhost:10082/

Show Users

    [root@localhost olv1]# curl -um64rxofsqkvlb4cj: -XGET http://localhost:10082/user

# Start Point on Linux

    [root@localhost olv1]# nohup ./resource/point -vs:addr x.x.x.x -vs:auth zzz:wwww -if:addr 192.168.x.b/24 &
    [root@localhost olv1]# ping 192.168.x.a

