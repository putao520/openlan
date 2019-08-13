# Overview 

Refer to https://github.com/danieldin95/openlan-py 

# Install tap-windows6

Download `resources/tap-windows-9.21.2.exe`, then install it. And click `cpe.exe` to run CPE in Windows. 

# Build in Powershell

    [root@localhost olan-v1.1]# go get github.com/songgao/water

    [root@localhost olan-v1.1]# go get github.com/milosgajdos83/tenus
    [root@localhost olan-v1.1]# go get golang.org/x/sys

    [root@localhost olan-v1.1]# go build -o cpe.exe cpe.go


# Configure Windows TAP Device

Goto `Control Panel\Network and Internet\Network Connections`, and find `Ethernet 2`, then you can configure IPAddress for it to access branch site. 


