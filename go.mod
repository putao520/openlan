module github.com/danieldin95/openlan

go 1.12

require (
	github.com/Sirupsen/logrus v0.11.0 // indirect
	github.com/akavel/rsrc v0.8.0 // indirect
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/coreos/go-systemd/v22 v22.0.0
	github.com/digitalocean/go-openvswitch v0.0.0-20210610204619-e0e0c7890c6b // indirect
	github.com/docker/libnetwork v0.5.6 // indirect
	github.com/go-ldap/ldap v3.0.3+incompatible
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/gorilla/mux v1.7.4
	github.com/moby/libnetwork v0.5.6
	github.com/shadowsocks/go-shadowsocks2 v0.1.5
	github.com/songgao/water v0.0.0-20190725173103-fd331bda3f4b
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.4 // indirect
	github.com/urfave/cli/v2 v2.3.0
	github.com/vishvananda/netlink v1.0.0
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df // indirect
	github.com/xtaci/kcp-go/v5 v5.5.12
	github.com/xtaci/kcptun v0.0.0-20200520151335-912a97993e20 // indirect
	golang.org/x/crypto v0.0.0 // indirect
	golang.org/x/net v0.0.0
	golang.org/x/sys v0.0.0 // indirect
	golang.org/x/time v0.0.0
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/yaml.v2 v2.2.3
)

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/net => github.com/golang/net v0.0.0-20190812203447-cdfb69ac37fc
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190209173611-3b5209105503
	golang.org/x/time => github.com/golang/time v0.0.0-20210220033141-f8bda1e9f3ba
)
