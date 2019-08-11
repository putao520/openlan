# Install tap-windows6

download `resources/windows/tap-windows-9.21.2.exe`, and install it. Next, click `./cpe.exe` to run CPE in Windows. 

# Build in Powershell

<<<<<<< HEAD
[root@localhost olan-v1.1]# go get github.com/songgao/water

[root@localhost olan-v1.1]# go get github.com/milosgajdos83/tenus
[root@localhost olan-v1.1]# go get golang.org/x/sys

[root@localhost olan-v1.1]# go build -o cpe.exe cpe_win.go
=======
    [root@localhost olan-v1.1]# go get github.com/songgao/water
    [root@localhost olan-v1.1]# go get github.com/milosgajdos83/tenus
    [root@localhost olan-v1.1]# go get golang.org/x/sys
    
    [root@localhost olan-v1.1]# go build -o cpe.exe cpe_win.go
>>>>>>> ca78111c802a1f25dc317967f74104744895116b

# Configure Windows TAP Device

goto `Control Panel\Network and Internet\Network Connections`, and find `Ethernet 2`, then you can configure IPAddress for it to access branch site. 


