package schema

type Route struct {
	Prefix  string `json:"prefix"`
	Nexthop string `json:"nexthop"`
}

type Network struct {
	Name    string  `json:"name"`
	IpStart string  `json:"ipStart"`
	IpEnd   string  `json:"ipEnd"`
	Netmask string  `json:"netmask"`
	Routes  []Route `json:"routes"`
}
