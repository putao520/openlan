package schema

type Lease struct {
	Address string `json:"address"`
	UUID    string `json:"uuid"`
	Alias   string `json:"alias"`
	Client  string `json:"client"`
	Type    string `json:"type"`
	Network string `json:"network"`
}

type PrefixRoute struct {
	Prefix  string `json:"prefix"`
	NextHop string `json:"nexthop"`
}

type Network struct {
	Name    string        `json:"name"`
	IpStart string        `json:"ipStart"`
	IpEnd   string        `json:"ipEnd"`
	Netmask string        `json:"netmask"`
	Routes  []PrefixRoute `json:"routes"`
}
