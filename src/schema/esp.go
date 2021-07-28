package schema

type Esp struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type EspState struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Auth   string `json:"auth"`
	Crypt  string `json:"crypt"`
}

type EspPolicy struct {
	Source string `json:"local"`
	Dest   string `json:"destination"`
}

type EspMember struct {
	Spi    int         `json:"spi"`
	Peer   string      `json:"peer"`
	State  EspState    `json:"state"`
	Policy []EspPolicy `json:"policy"`
}
