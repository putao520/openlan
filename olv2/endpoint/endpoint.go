package endpoint

import (
	"log"
)

type Endpoint struct {
	bridge *Bridge
	verbose bool
}

func NewEndpoint(c *Config) (this *Endpoint) {
	this = &Endpoint {
		bridge: NewBridge(c),
		verbose: c.Verbose,
	}

	return
}

func (this *Endpoint) Start() {
	log.Printf("Info| Endpoint.Start")
	this.bridge.Start()
}

func (this *Endpoint) Stop() {
	log.Printf("Info| Endpoint.Stop")
	this.bridge.Stop()
}