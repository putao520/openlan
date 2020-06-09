package schema

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
