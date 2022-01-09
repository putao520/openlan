module github.com/danieldin95/openlan

go 1.12

require (
	github.com/Sirupsen/logrus v0.11.0 // indirect
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/chzyer/logex v1.2.0 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/chzyer/test v0.0.0-20210722231415-061457976a23 // indirect
	github.com/coreos/etcd v3.3.27+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/go-systemd/v22 v22.0.0
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/danieldin95/go-openvswitch v0.0.5
	github.com/docker/libnetwork v0.5.6 // indirect
	github.com/go-ldap/ldap v3.0.3+incompatible
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/klauspost/cpuid v1.2.3 // indirect
	github.com/moby/libnetwork v0.5.6
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shadowsocks/go-shadowsocks2 v0.1.5
	github.com/songgao/water v0.0.0-20190725173103-fd331bda3f4b
	github.com/stretchr/testify v1.7.0
	github.com/tjfoc/gmsm v1.3.0 // indirect
	github.com/urfave/cli/v2 v2.3.0
	github.com/vishvananda/netlink v1.0.0
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df // indirect
	github.com/xtaci/kcp-go/v5 v5.5.12
	go.uber.org/zap v1.20.0 // indirect
	golang.org/x/crypto v0.0.0 // indirect
	golang.org/x/net v0.0.0
	golang.org/x/sys v0.0.0 // indirect
	golang.org/x/time v0.0.0
	google.golang.org/grpc v1.26.0 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/yaml.v2 v2.2.8
)

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/net => github.com/golang/net v0.0.0-20190812203447-cdfb69ac37fc
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190209173611-3b5209105503
	golang.org/x/time => github.com/golang/time v0.0.0-20210220033141-f8bda1e9f3ba
)
