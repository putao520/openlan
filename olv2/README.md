# Overview 

Refer to `Overview.md`

# Install tap-windows6

Download `resources/tap-windows-9.21.2.exe`, then install it. And run CPE in Windows by `endpoint.exe -addr x.x.x.x:10020 -auth zzz@nnn:wwww`. 

# Build in Powershell

Download dependent sources

    PS L:\openlan-go\olv2> go get github.com/songgao/water
    PS L:\openlan-go\olv2> go get github.com/milosgajdos83/tenus
    PS L:\openlan-go\olv2> go get golang.org/x/sys

The following command for you building endpoint on windows

    PS L:\openlan-go\olv2> go build -o ./resources/endpoint.exe endpoint_windows.go

# Configure Windows TAP Device

Goto `Control Panel\Network and Internet\Network Connections`, and find `Ethernet 2`, then you can configure IPAddress like `192.168.x.a` for other branch site to access it. 

Setting MTU, run `cmd` by administrator priviledge. 

    netsh interface ipv4 show subinterfaces
    netsh interface ipv4 set subinterface "Ethernet 2" mtu=1430 store=persistent

# Start Controller on Linux

    [root@localhost olv1]# cat .passowrd
    zzz@nnn:wwww
    xxxx@ooo:aaaaa
    [root@localhost olv1]# nohup ./resources/endpoint -addr x.x.x.x:10020 &
    [root@localhost olv1]# cat .opetoken
    m64rxofsqkvlb4cj

You can use `curl` to show CPEs already onlines.

    [root@localhost olv1]# curl -um64rxofsqkvlb4cj: -XGET http://localhost:10082/

And search or show users can login.

    [root@localhost olv1]# curl -um64rxofsqkvlb4cj: -XGET http://localhost:10082/user

# Start Endpoint on Linux

    [root@localhost olv1]# nohup ./resources/endpoint -addr x.x.x.x:10002 -auth zzz@nnn:wwww &
    [root@localhost olv1]# ifconfig tap0 192.168.x.b up
    [root@localhost olv1]# ping 192.168.x.a
