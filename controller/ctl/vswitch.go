package ctl

import (
	"github.com/danieldin95/lightstar/libstar"
	"github.com/danieldin95/openlan-go/controller/schema"
	"github.com/danieldin95/openlan-go/libol"
	"sync"
	"time"
)

type VSwitch struct {
	Lock      sync.RWMutex
	Service   schema.VSwitch
	Ticker    *time.Ticker
	Done      chan bool
	Points    []schema.Point
	Links     []schema.Point
	Neighbors []schema.Neighbor
	Auth      libstar.Auth
	State     string
}

func (v *VSwitch) Init() {
	v.Auth = libstar.Auth{
		Type:     "basic",
		Username: v.Service.Password + ":",
	}
	if v.Ticker == nil {
		v.Ticker = time.NewTicker(5 * time.Second)
	}
	if v.Done == nil {
		v.Done = make(chan bool)
	}
}

func (v *VSwitch) Once() error {
	libol.Debug("VSwitch.Once %s", v)
	// point
	client := libstar.HttpClient{
		Url:  v.Service.Url + "/api/point",
		Auth: v.Auth,
	}
	resp, err := client.Do()
	if err != nil {
		return err
	}
	if err := libstar.GetJSON(resp.Body, &v.Points); err != nil {
		return err
	}
	libol.Debug("VSwitch.Once %s", v.Points)

	// link
	client = libstar.HttpClient{
		Url:  v.Service.Url + "/api/link",
		Auth: v.Auth,
	}
	resp, err = client.Do()
	if err != nil {
		return err
	}
	if err := libstar.GetJSON(resp.Body, &v.Links); err != nil {
		return err
	}
	libol.Debug("VSwitch.Once %s", v.Links)

	// neighbor
	client = libstar.HttpClient{
		Url:  v.Service.Url + "/api/neighbor",
		Auth: v.Auth,
	}
	resp, err = client.Do()
	if err != nil {
		return err
	}
	if err := libstar.GetJSON(resp.Body, &v.Neighbors); err != nil {
		return err
	}
	libol.Debug("VSwitch.Once %s", v.Neighbors)
	v.State = ""

	return nil
}

func (v *VSwitch) Start() {
	libol.Info("VSwitch.Start %s", v)
	go func() {
		for {
			select {
			case <-v.Done:
				return
			case <-v.Ticker.C:
				if err := v.Once(); err != nil {
					if v.State != err.Error() {
						v.State = err.Error()
						libol.Error("VSwitch.Ticker %s %s", v, err)
					}
				}
			}
		}
	}()
}

func (v *VSwitch) Stop() {
	libol.Info("VSwitch.Stop %s", v)
	v.Done <- true
}

func (v *VSwitch) String() string {
	return v.Service.Name
}

func (v *VSwitch) ListPoint() <-chan *schema.Point {
	c := make(chan *schema.Point, 128)
	go func() {
		v.Lock.RLock()
		defer v.Lock.RUnlock()

		for i := range v.Points {
			c <- &v.Points[i]
		}
		c <- nil // Finish channel by nil.
	}()
	return c
}

func (v *VSwitch) ListLink() <-chan *schema.Point {
	c := make(chan *schema.Point, 128)
	go func() {
		v.Lock.RLock()
		defer v.Lock.RUnlock()

		for i := range v.Links {
			c <- &v.Links[i]
		}
		c <- nil // Finish channel by nil.
	}()
	return c
}

func (v *VSwitch) ListNeighbor() <-chan *schema.Neighbor {
	c := make(chan *schema.Neighbor, 128)
	go func() {
		v.Lock.RLock()
		defer v.Lock.RUnlock()

		for i := range v.Neighbors {
			c <- &v.Neighbors[i]
		}
		c <- nil // Finish channel by nil.
	}()
	return c
}
