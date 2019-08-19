package endpoint

import (
	"log"

	"github.com/danieldin95/openlan-go/olv2/openlanv2"
)

type Endpoint struct {
	bridge *Bridge
	verbose bool
	http *Http
}

func NewEndpoint(c *Config) (this *Endpoint) {
	this = &Endpoint {
		bridge: NewBridge(c),
		verbose: c.Verbose,
	}

	this.http = NewHttp(this, c)

	return
}

func (this *Endpoint) Start() {
	log.Printf("Info| Endpoint.Start")
	this.bridge.Start()
	go this.http.GoStart()
}

func (this *Endpoint) Stop() {
	log.Printf("Info| Endpoint.Stop")
	this.bridge.Stop()
}

func (this *Endpoint) GetPeers() chan *openlanv2.Endpoint {
    return this.bridge.Network.ListEndpoint()
}

func (this *Endpoint) GetMacs() chan *MacEntry {
    return this.bridge.ListMacs()
}