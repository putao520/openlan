package olsw

import (
	co "github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olsw/api"
	nl "github.com/vishvananda/netlink"
	"net"
)

type VxLANWorker struct {
	uuid  string
	cfg   *co.Network
	inCfg *co.VxLANInterface
	out   *libol.SubLogger
}

func NewVxLANWorker(c *co.Network) *VxLANWorker {
	w := &VxLANWorker{
		cfg: c,
		out: libol.NewSubLogger(c.Name),
	}
	w.inCfg, _ = c.Interface.(*co.VxLANInterface)
	return w
}

func (w *VxLANWorker) Initialize() {
}

func (w *VxLANWorker) UpBr(name string) (*nl.Bridge, error) {
	la := nl.LinkAttrs{TxQLen: -1, Name: name}
	br := &nl.Bridge{LinkAttrs: la}
	if link, err := nl.LinkByName(name); link == nil {
		w.out.Warn("VxLANWorker.UpBr: %s %s", name, err)
		if err := nl.LinkAdd(br); err != nil {
			return nil, err
		}
	}
	if link, err := nl.LinkByName(name); link == nil {
		return nil, err
	} else if err := nl.LinkSetUp(link); err != nil {
		w.out.Warn("VxLANWorker.UpBr %s", err)
	}
	return br, nil
}

func (w *VxLANWorker) UpVxLAN(cfg *co.VxLANMember) error {
	name := cfg.Name
	link, _ := nl.LinkByName(name)
	if link == nil {
		port := &nl.Vxlan{
			LinkAttrs: nl.LinkAttrs{
				TxQLen: -1,
				Name:   name,
			},
			VxlanId: cfg.VNI,
			SrcAddr: net.ParseIP(cfg.Local),
			Group:   net.ParseIP(cfg.Remote),
			Port:    cfg.Port,
		}
		if err := nl.LinkAdd(port); err != nil {
			return err
		}
		link, _ = nl.LinkByName(name)
	}
	if err := nl.LinkSetUp(link); err != nil {
		w.out.Error("VxLANWorker.UpVxLAN: %s", err)
	}
	if br, err := w.UpBr(cfg.Bridge); err != nil {
		return err
	} else if err := nl.LinkSetMaster(link, br); err != nil {
		return err
	}
	return nil
}

func (w *VxLANWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	if w.inCfg == nil {
		w.out.Error("VxLANWorker.Start inCfg is nil")
		return
	}
	for _, mem := range w.inCfg.Members {
		if err := w.UpVxLAN(mem); err != nil {
			w.out.Error("VxLANWorker.Start %s %s", mem.Name, err)
		}
	}
}

func (w *VxLANWorker) DownVxLAN(cfg *co.VxLANMember) error {
	name := cfg.Name
	link, _ := nl.LinkByName(name)
	if link == nil {
		return nil
	}
	port := &nl.Vxlan{
		LinkAttrs: nl.LinkAttrs{
			TxQLen: -1,
			Name:   name,
		},
	}
	if err := nl.LinkDel(port); err != nil {
		return err
	}
	return nil
}

func (w *VxLANWorker) Stop() {
	if w.inCfg == nil {
		w.out.Error("VxLANWorker.Stop inCfg is nil")
		return
	}
	for _, mem := range w.inCfg.Members {
		if err := w.DownVxLAN(mem); err != nil {
			w.out.Error("VxLANWorker.Stop %s %s", mem.Name, err)
		}
	}
}

func (w *VxLANWorker) String() string {
	return w.cfg.Name
}

func (w *VxLANWorker) ID() string {
	return w.uuid
}

func (w *VxLANWorker) GetBridge() network.Bridger {
	w.out.Warn("VxLANWorker.GetBridge operation notSupport")
	return nil
}

func (w *VxLANWorker) GetConfig() *co.Network {
	return w.cfg
}

func (w *VxLANWorker) GetSubnet() string {
	w.out.Warn("VxLANWorker.GetSubnet operation notSupport")
	return ""
}
