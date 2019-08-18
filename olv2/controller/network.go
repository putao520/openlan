package controller

type Network struct {
	Name string
	Endpoints map[string]*Endpoint // UUID is key
}

func NewNetwork(name string)(this *Network) {
	this = &Network {
		Name: name,
		Endpoints: make(map[string]*Endpoint),
	}
	return
}

func (this *Network) AddEndpoint(point *Endpoint) error {
	if _, ok := this.Endpoints[point.UUID]; ok {
		return nil
	}

	this.Endpoints[point.UUID] = point
	return nil
}

func (this *Network) DelEndpoint(uuid string) error {
	if _, ok := this.Endpoints[uuid]; ok {
		delete(this.Endpoints, uuid)
	}
	return nil
}