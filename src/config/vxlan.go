package config

type VxLANMember struct {
	Name    string `json:"name"`
	VNI     int    `json:"vni"`
	Local   string `json:"local"`
	Remote  string `json:"remote"`
	Network string `json:"network"`
	Bridge  string `json:"bridge"`
}

type VxLANInterface struct {
	Name    string        `json:"name"`
	Local   string        `json:"local"`
	Members []VxLANMember `json:"members"`
}
