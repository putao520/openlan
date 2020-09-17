package schema

type Device struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Mac      string `json:"mac"`
	Type     string `json:"type"`
	Provider string `json:"provider"`
	Mtu      int    `json:"mtu"`
}
