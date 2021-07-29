package schema

type Esp struct {
	Name    string      `json:"name"`
	Address string      `json:"address"`
	Members []EspMember `json:"members,omitempty"`
}

type EspState struct {
	Name       string `json:"name"`
	Spi        int    `json:"spi"`
	Source     string `json:"source"`
	Mode       uint8  `json:"mode"`
	Proto      uint8  `json:"proto"`
	Dest       string `json:"destination"`
	Auth       string `json:"auth"`
	Crypt      string `json:"crypt"`
	TxBytes    uint64 `json:"txBytes"`
	TxPackages uint64 `json:"txPackages"`
	RxBytes    uint64 `json:"rxBytes"`
	RxPackages uint64 `json:"rxPackages"`
}

type EspPolicy struct {
	Name   string `json:"name"`
	Spi    int    `json:"spi"`
	Source string `json:"local"`
	Dest   string `json:"destination"`
}

type EspMember struct {
	Name   string      `json:"name"`
	Spi    int         `json:"spi"`
	Peer   string      `json:"peer"`
	State  EspState    `json:"state"`
	Policy []EspPolicy `json:"policy"`
}
