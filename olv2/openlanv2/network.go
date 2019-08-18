package openlanv2

type Network struct {
	Name string
	Endpoints map[string]*Endpoint // default UUID is key
}

func NewNetwork(name string)(this *Network) {
	this = &Network {
		Name: name,
		Endpoints: make(map[string]*Endpoint),
	}
	return
}

func (this *Network) AddEndpoint(uuid string, point *Endpoint) error {
	if _, ok := this.Endpoints[uuid]; ok {
		return nil
	}

	this.Endpoints[uuid] = point
	return nil
}

func (this *Network) DelEndpoint(uuid string) error {
	if _, ok := this.Endpoints[uuid]; ok {
		delete(this.Endpoints, uuid)
	}
	return nil
}