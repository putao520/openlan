# Install tap-windows6 firstlly 

download `resources/windows/tap-windows-9.21.2.exe`, and install it.

# Run Binary Redirectly
click `./cpe.exe` 

# Build in Powershell

    [root@localhost olan-v1.1]# go get github.com/songgao/water
    [root@localhost olan-v1.1]# go get github.com/milosgajdos83/tenus
    [root@localhost olan-v1.1]# go get golang.org/x/sys
    
    [root@localhost olan-v1.1]# go build -o cpe.exe cpe_win.go`

# Configure Windows TAP Device

goto `Control Panel\Network and Internet\Network Connections`, and find `Ethernet 2`, then you can configure IPAddress for it to access branch site. 


