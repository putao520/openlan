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

func (c *ConfD) Add(table string, model model.Model) {
	if obj, ok := model.(*database.Switch); ok {
		c.out.Info("ConfD.Add switch %d", obj.Listen)
	}

	if obj, ok := model.(*database.VirtualNetwork); ok {
		c.out.Info("ConfD.Add virtual network %s %s", obj.Name, obj.Address)
	}

	if obj, ok := model.(*database.VirtualLink); ok {
		c.out.Info("ConfD.Add virtual link %s %s", obj.Network, obj.Connection)
		proto := libol.GetPrefix(obj.Connection, 4)
		if proto == "spi:" || proto == "udp:" {
			c.AddMember(obj)
		}
	}

	if obj, ok := model.(*database.NameCache); ok {
		c.UpdateName(obj)
	}
}

func (c *ConfD) Delete(table string, model model.Model) {
	if obj, ok := model.(*database.VirtualNetwork); ok {
		c.out.Info("ConfD.Delete virtual network %s %s", obj.Name, obj.Address)
	}

	if obj, ok := model.(*database.VirtualLink); ok {
		c.out.Info("ConfD.Delete virtual link %s %s", obj.Network, obj.Connection)
	}
}

func (c *ConfD) Update(table string, old model.Model, new model.Model) {
	if obj, ok := new.(*database.VirtualNetwork); ok {
		c.out.Info("ConfD.Update virtual network %s %s", obj.Name, obj.Address)
	}

	if obj, ok := new.(*database.VirtualLink); ok {
		c.out.Info("ConfD.Update virtual link %s %s", obj.Network, obj.Connection)
		proto := libol.GetPrefix(obj.Connection, 4)
		if proto == "spi:" || proto == "udp:" {
			c.AddMember(obj)
		}
	}

	if obj, ok := new.(*database.NameCache); ok {
		c.UpdateName(obj)
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
	if libol.GetPrefix(conn, 4) == "spi:" {
		remoteConn := obj.Status["remote_connection"]
		if libol.GetPrefix(remoteConn, 4) == "udp:" {
			spi, _ = strconv.Atoi(conn[4:])
			remote, port = GetAddrPort(remoteConn[4:])
		} else {
			c.out.Warn("ConfD.AddMember %s remote not found.", conn)
			return
		}
	} else if libol.GetPrefix(conn, 4) == "udp:" {
		remoteConn := obj.Connection
		spi, _ = strconv.Atoi(libol.GetSuffix(obj.Device, 3))
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
			Auth:       obj.Authentication["password"],
			Crypt:      obj.Authentication["username"],
		},
	}
	c.out.Info("ConfD.AddMember %v", memCfg)
	worker := GetWorker(obj.Network)
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
		worker.Reload(netCfg)
	}
}

func (c *ConfD) UpdateName(obj *database.NameCache) {
	if obj.Address == "" {
		return
	}
	c.out.Info("ConfD.UpdateName %s %s", obj.Name, obj.Address)
	ListWorker(func(w Networker) {
		cfg := w.GetConfig()
		if cfg.Provider != "esp" {
			return
		}
		spec := cfg.Specifies
		found := false
		if specObj, ok := spec.(*config.ESPSpecifies); ok {
			for _, mem := range specObj.Members {
				state := mem.State
				if state.Remote != obj.Name {
					continue
				}
				if state.RemoteIp.String() == obj.Address {
					continue
				}
				found = true
			}
		}
		if found {
			cfg.Correct()
			w.Reload(cfg)
		}
	})
}
