package openlanv2

import (
	"sync"
	//"log"
)

type Network struct {
	Name string
	Endpoints map[string]*Endpoint // default UUID is key
	//
	rwlock sync.RWMutex
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

	this.rwlock.Lock()
	defer this.rwlock.Unlock()
	this.Endpoints[uuid] = point
	return nil
}

func (this *Network) DelEndpoint(uuid string) error {
	this.rwlock.Lock()
	defer this.rwlock.Unlock()
	if _, ok := this.Endpoints[uuid]; ok {
		delete(this.Endpoints, uuid)
	}
	return nil
}

func (this *Network) GetEndpoint(uuid string) (*Endpoint) {
	this.rwlock.RLock()
	defer this.rwlock.RUnlock()
	if point, ok := this.Endpoints[uuid]; ok {
		return point
	}
	return nil
}

func (this *Network) ListEndpoint() chan *Endpoint {
	c := make(chan *Endpoint, 16)
    go func() {
		this.rwlock.RLock()
		defer this.rwlock.RUnlock()

        for _, peer := range this.Endpoints {
			//log.Printf("Debug| Endpoint.GetPeers: %s", peer)
            c <- peer
		}
		c <- nil //Finish channel by nil.
    }()

    return c
}
