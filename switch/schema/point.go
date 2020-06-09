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
