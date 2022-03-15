package olsw

import (
	"github.com/danieldin95/openlan/pkg/config"
	"github.com/danieldin95/openlan/pkg/database"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/cache"
	"github.com/ovn-org/libovsdb/model"
	"strconv"
	"strings"
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
	handler := &cache.EventHandlerFuncs{
		AddFunc:    c.Add,
		DeleteFunc: c.Delete,
		UpdateFunc: c.Update,
	}
	if _, err := database.NewDBClient(handler); err != nil {
		c.out.Error("Confd.Start open db with %s", err)
		return
	}
}

func (c *ConfD) Stop() {
}

func GetPrefix(value string, index int) string {
	if len(value) >= index {
		return value[:index]
	}
	return ""
}

func GetSuffix(value string, index int) string {
	if len(value) >= index {
		return value[index:]
	}
	return ""
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
		proto := GetPrefix(obj.Connection, 4)
		if proto == "spi:" || proto == "udp:" {
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
		proto := GetPrefix(obj.Connection, 4)
		if proto == "spi:" || proto == "udp:" {
			c.AddMember(obj)
		}
	}
}

func GetAddrPort(conn string) (string, int) {
	addrs := strings.SplitN(conn, ":", 2)
	if len(addrs) == 2 {
		port, _ := strconv.Atoi(addrs[1])
		return addrs[0], port
	}
	return addrs[0], 0
}

func (c *ConfD) AddMember(obj *database.VirtualLink) {
	var spi, port int
	var remote, remoteConn string
	conn := obj.Connection
	if GetPrefix(conn, 4) == "spi:" {
		remoteConn := obj.Status["remote_connection"]
		if GetPrefix(remoteConn, 4) == "udp:" {
			spi, _ = strconv.Atoi(conn[4:])
			remote, port = GetAddrPort(remoteConn[4:])
		} else {
			c.out.Warn("ConfD.AddMember %s remote not found.", conn)
			return
		}
	} else if GetPrefix(conn, 4) == "udp:" {
		remoteConn := obj.Connection
		spi, _ = strconv.Atoi(GetSuffix(obj.Device, 3))
		remote, port = GetAddrPort(remoteConn[4:])
	} else {
		return
	}
	c.out.Info("ConfD.AddMember remote link %s %s", conn, remoteConn)
	memCfg := &config.ESPMember{
		Name:    obj.Device,
		Address: obj.OtherConfig["local_address"],
		Peer:    obj.OtherConfig["remote_address"],
		Spi:     spi,
		State: config.EspState{
			Remote:     remote,
			RemotePort: port,
		},
	}
	c.out.Info("ConfD.AddMember %v", memCfg)
	worker := Workers[obj.Network]
	if worker == nil {
		c.out.Warn("ConfD.AddMember network %s not found.", obj.Network)
		return
	}
	netCfg := worker.GetConfig()
	if netCfg == nil {
		c.out.Warn("ConfD.AddMember config %s not found.", obj.Network)
		return
	}
	if netCfg.Provider == "esp" {
		spec := netCfg.Specifies
		if specObj, ok := spec.(*config.ESPSpecifies); ok {
			found := false
			for index, mem := range specObj.Members {
				if mem.Spi == memCfg.Spi {
					found = true
					specObj.Members[index] = memCfg
				}
			}
			if !found {
				specObj.Members = append(specObj.Members, memCfg)
			}
			specObj.Correct()
		}
		worker.Reload(nil)
	}
}
