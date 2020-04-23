package schema

type Point struct {
	Uptime  int64  `json:"uptime"`
	UUID    string `json:"uuid"`
	Network string `json:"network"`
	Alias   string `json:"alias"`
	Address string `json:"server"`
	Switch  string `json:"switch"`
	IpAddr  string `json:"address"`
	Device  string `json:"device"`
	RxBytes uint64 `json:"rxBytes"`
	TxBytes uint64 `json:"txBytes"`
	ErrPkt  uint64 `json:"errors"`
	State   string `json:"state"`
}

type Link struct {
	Uptime  int64  `json:"uptime"`
	UUID    string `json:"uuid"`
	Alias   string `json:"alias"`
	Network string `json:"network"`
	Address string `json:"server"`
	IpAddr  string `json:"address"`
	Device  string `json:"device"`
	RxBytes uint64 `json:"rxBytes"`
	TxBytes uint64 `json:"txBytes"`
	ErrPkt  uint64 `json:"errors"`
	State   string `json:"state"`
}

type LinkConfig struct {
	Alias   string `json:"alias"`
	Addr    string `json:"vs.addr"`
	Auth    string `json:"vs.auth"`
	Tls     bool   `json:"vs.tls"`
	Allowed bool   `json:"vs.allowed"`
	IfName  string `json:"if.name"`
	IfMtu   int    `json:"if.mtu"`
	IfAddr  string `json:"if.addr"`
	BrName  string `json:"if.br"`
	IfTun   bool   `json:"if.tun"`
	LogFile string `json:"log.file"`
	Verbose int    `json:"log.level"`
	Script  string `json:"script"`
	Network string `json:"network"`
}

type Neighbor struct {
	Uptime int64  `json:"uptime"`
	UUID   string `json:"uuid"`
	HwAddr string `json:"ethernet"`
	IpAddr string `json:"address"`
	Client string `json:"client"`
}

type OnLine struct {
	Uptime     int64  `json:"uptime"`
	EthType    uint16 `json:"ethType"`
	IpSource   string `json:"ipSource"`
	IpDest     string `json:"ipDestination"`
	IpProto    string `json:"ipProtocol"`
	PortSource uint16 `json:"portSource"`
	PortDest   uint16 `json:"portDestination"`
}

type Index struct {
	Version   Version    `json:"version"`
	Worker    Worker     `json:"worker"`
	Points    []Point    `json:"points"`
	Links     []Link     `json:"links"`
	Neighbors []Neighbor `json:"neighbors"`
	OnLines   []OnLine   `json:"online"`
	Network   []Network  `json:"network"`
}

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Alias    string `json:"alias"`
}

type Route struct {
	Prefix  string `json:"prefix"`
	Nexthop string `json:"nexthop"`
}

type Network struct {
	Name    string  `json:"name"`
	IfAddr  string  `json:"ifAddr"`
	IpAddr  string  `json:"ipAddr"`
	IpRange int     `json:"ipRange"`
	Netmask string  `json:"netmask"`
	Routes  []Route `json:"routes"`
}

type Ctrl struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}
