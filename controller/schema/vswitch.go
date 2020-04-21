package schema

import "net/url"

type VSwitch struct {
	Name     string      `json:"name"`
	Url      string      `json:"url"`
	Schema   string      `json:"schema"`
	Address  string      `json:"address"`
	Password string      `json:"password"`
	State    string      `json:"state"`
	Ctl      interface{} `json:"-"`
}

func (v *VSwitch) Init() {
	if u, err := url.Parse(v.Url); err == nil {
		v.Address = u.Host
		v.Schema = u.Scheme
		if u.Port() == "" {
			v.Url += ":10000"
			v.Address += ":10000"
		}
	}
}
