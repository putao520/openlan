package olsw

import (
	"github.com/danieldin95/openlan/pkg/config"
	"github.com/danieldin95/openlan/pkg/database"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/cache"
	"github.com/ovn-org/libovsdb/model"
	"strconv"
)

type ConfD struct {
	stop chan struct{}
	out  *libol.SubLogger
}

func NewConfd() *ConfD {
	c := &ConfD{
		out:  libol.NewSubLogger("confd"),
		stop: make(chan struct{}),
	}
	return c
}

func (c *ConfD) Initialize() {
}

func (c *ConfD) Start() {
	if _, err := database.NewDBClient(); err != nil {
		c.out.Error("Confd.Start open db with %s", err)
		return
	}
	handler := &cache.EventHandlerFuncs{
		AddFunc:    c.Add,
		DeleteFunc: c.Delete,
		UpdateFunc: c.Update,
	}
	processor := database.Client.Client.Cache()
	processor.AddEventHandler(handler)
	processor.Run(c.stop)
}

func (c *ConfD) Stop() {
	c.stop <- struct{}{}
}

func (c *ConfD) Add(table string, model model.Model) {
	if table == "Global_Switch" {
		obj := model.(*database.Switch)
		c.out.Info("ConfD.Add switch %d", obj.Listen)
	}
	if table == "Virtual_Network" {
		obj := model.(*database.VirtualNetwork)
		c.out.Info("ConfD.Add virtual network %s %s", obj.Name, obj.Address)
	}
	if table == "Virtual_Link" {
		obj := model.(*database.VirtualLink)
		c.out.Info("ConfD.Add virtual link %s %s", obj.Network, obj.Connection)
		if obj.Connection[:4] == "spi:" {
			c.AddMember(obj)
		}
	}
}

func (c *ConfD) Delete(table string, model model.Model) {
	if table == "Virtual_Network" {
		obj := model.(*database.VirtualNetwork)
		c.out.Info("ConfD.Delete virtual network %s %s", obj.Name, obj.Address)
	}
	if table == "Virtual_Link" {
		obj := model.(*database.VirtualLink)
		c.out.Info("ConfD.Delete virtual link %s %s", obj.Network, obj.Connection)
	}
}

func (c *ConfD) Update(table string, old model.Model, new model.Model) {
	if table == "Virtual_Network" {
		obj := new.(*database.VirtualNetwork)
		c.out.Info("ConfD.Update virtual network %s %s", obj.Name, obj.Address)
	}
	if table == "Virtual_Link" {
		obj := new.(*database.VirtualLink)
		c.out.Info("ConfD.Update virtual link %s %s", obj.Network, obj.Connection)
		if obj.Connection[:4] == "spi:" {
			c.AddMember(obj)
		}
	}
}

func (c *ConfD) AddMember(obj *database.VirtualLink) {
	remoteConn := obj.Status["remote_connection"]
	spi, _ := strconv.Atoi(obj.Connection[4:])
	c.out.Info("ConfD.AddMember remote link %s %s", obj.Connection, remoteConn)
	memCfg := config.ESPMember{
		Name:    obj.Network,
		Address: obj.OtherConfig["local_address"],
		Peer:    obj.OtherConfig["remote_address"],
		Spi:     spi,
		State: config.EspState{
			Remote: remoteConn[4:],
		},
	}
	c.out.Info("ConfD.AddMember %v", memCfg)
	//TODO update to esp configuration.
}
