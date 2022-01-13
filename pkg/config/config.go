package config

type manager struct {
	Point  *Point
	Switch *Switch
	Proxy  *Proxy
}

var Manager = manager{
	Point:  &Point{},
	Switch: &Switch{},
	Proxy:  &Proxy{},
}
