package schema

type Index struct {
	Version   Version    `json:"version"`
	Worker    Worker     `json:"worker"`
	Points    []Point    `json:"points"`
	Links     []Link     `json:"links"`
	Neighbors []Neighbor `json:"neighbors"`
	OnLines   []OnLine   `json:"online"`
	Network   []Network  `json:"network"`
	OvClients []OvClient `json:"ovclients"`
}

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Alias    string `json:"alias"`
	Network  string `json:"network"`
}

type Ctrl struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}
