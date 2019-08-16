# Overview 

Refer to https://github.com/danieldin95/openlan-py 

# Install tap-windows6

Download `resources/tap-windows-9.21.2.exe`, then install it. And run CPE in Windows by `cpe.exe -addr x.x.x.x:10002 -auth zzz:wwww`. 

# Build in Powershell

Download dependent sources

    PS L:\openlan-go\olv1> go get github.com/songgao/water
    PS L:\openlan-go\olv1> go get github.com/milosgajdos83/tenus
    PS L:\openlan-go\olv1> go get golang.org/x/sys

Execute building command

    PS L:\openlan-go\olv1> go build -o ./resources/cpe.exe cpe_windows.go

# Configure Windows TAP Device

Goto `Control Panel\Network and Internet\Network Connections`, and find `Ethernet 2`, then you can configure IPAddress for it to access branch site. 

# Start OPE on Linux

    [root@localhost olv1]# cat .passowrd
    zzz:wwww
    xxxx:aaaaa
    [root@localhost olv1]# nohup ./resources/ope -addr x.x.x.x:10002 &
    [root@localhost olv1]# cat .opetoken
    m64rxofsqkvlb4cj
    [root@localhost olv1]# ifconfig br-olan-10002 192.168.x.a up
    
Show CPEs

    [root@localhost olv1]# curl -um64rxofsqkvlb4cj: -XGET http://localhost:10082/

Show Users

    [root@localhost olv1]# curl -um64rxofsqkvlb4cj: -XGET http://localhost:10082/user

# Start CPE on Linux

    [root@localhost olv1]# nohup ./resources/cpe -addr x.x.x.x:10002 -auth zzz:wwww &
    [root@localhost olv1]# ifconfig tap0 192.168.x.b up
    [root@localhost olv1]# ping 192.168.x.a

