package config

type ESPState struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Auth   string `json:"auth"`
	Secret string `json:"secret"`
}

type ESPPolicy struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type ESPMember struct {
	Name     string      `json:"name"`
	Local    string      `json:"local"`
	Remote   string      `json:"remote"`
	Spi      int         `json:"spi"`
	State    ESPState    `json:"state"`
	Policies []ESPPolicy `json:"policies"`
}

type ESPInterface struct {
	Name    string      `json:"name"`
	Spi     int         `json:"spi"`
	Local   string      `json:"local"`
	State   ESPState    `json:"state"`
	Members []ESPMember `json:"members"`
}
