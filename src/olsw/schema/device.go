package schema

type Device struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Mac      string `json:"mac"`
	Type     string `json:"type"`
	Provider string `json:"provider"`
	Mtu      int    `json:"mtu"`
}

type HwMacInfo struct {
	Uptime  int64  `json:"uptime"`
	Address string `json:"address"`
	Device  string `json:"device"`
}

type Bridge struct {
	Device
	Macs []HwMacInfo `json:"macs"`
}
