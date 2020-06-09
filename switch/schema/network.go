package schema

type Lease struct {
	Address string `json:"address"`
	UUID    string `json:"uuid"`
	Client  string `json:"client"`
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
