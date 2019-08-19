package openlanv2

import (
	"sync"
)

type Network struct {
	Name string
	Endpoints map[string]*Endpoint // default UUID is key
	EndpointsRWLock sync.RWMutex
}

func NewNetwork(name string)(this *Network) {
	this = &Network {
		Name: name,
		Endpoints: make(map[string]*Endpoint),
	}
	return
}

func (this *Network) AddEndpoint(uuid string, point *Endpoint) error {
	if _point := this.GetEndpoint(uuid); _point != nil {
		return nil
	}

	this.EndpointsRWLock.Lock()
	defer this.EndpointsRWLock.Unlock()
	this.Endpoints[uuid] = point
	return nil
}

func (this *Network) DelEndpoint(uuid string) error {
	this.EndpointsRWLock.Lock()
	defer this.EndpointsRWLock.Unlock()
	if _, ok := this.Endpoints[uuid]; ok {
		delete(this.Endpoints, uuid)
	}
	return nil
}

func (this *Network) GetEndpoint(uuid string) (*Endpoint) {
	this.EndpointsRWLock.RLock()
	defer this.EndpointsRWLock.RUnlock()
	if point, ok := this.Endpoints[uuid]; ok {
		return point
	}
	return nil
}