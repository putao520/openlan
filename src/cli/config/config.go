package config

type _manager struct {
	Point  *Point
	Switch *Switch
	Proxy  *Proxy
}

var Manager = _manager{}
