package olsw

import (
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olsw/api"
	"github.com/vishvananda/netlink"
	"net"
)

type VxLANWorker struct {
	uuid  string
	cfg   *config.Network
	inCfg *config.VxLANInterface
	out   *libol.SubLogger
}

func NewVxLANWorker(c *config.Network) *VxLANWorker {
	w := &VxLANWorker{
		cfg: c,
		out: libol.NewSubLogger(c.Name),
	}
	w.inCfg, _ = c.Interface.(*config.VxLANInterface)
	return w
}

func (w *VxLANWorker) Initialize() {
}

func (w *VxLANWorker) UpBr(name string) (*netlink.Bridge, error) {
	la := netlink.LinkAttrs{TxQLen: -1, Name: name}
	br := &netlink.Bridge{LinkAttrs: la}
	if link, err := netlink.LinkByName(name); link == nil {
		w.out.Warn("VxLANWorker.UpBr: %s %s", name, err)
		if err := netlink.LinkAdd(br); err != nil {
			return nil, err
		}
	}
	if link, err := netlink.LinkByName(name); link == nil {
		return nil, err
	} else if err := netlink.LinkSetUp(link); err != nil {
		w.out.Warn("VxLANWorker.UpBr %s", err)
	}
	return br, nil
}

func (w *VxLANWorker) UpVxLAN(cfg *config.VxLANMember) error {
	name := cfg.Name
	link, _ := netlink.LinkByName(name)
	if link == nil {
		port := &netlink.Vxlan{
			LinkAttrs: netlink.LinkAttrs{
				TxQLen: -1,
				Name:   name,
			},
			VxlanId: cfg.VNI,
			SrcAddr: net.ParseIP(cfg.Local),
			Group:   net.ParseIP(cfg.Remote),
			Port:    cfg.Port,
		}
		if err := netlink.LinkAdd(port); err != nil {
			return err
		}
		link, _ = netlink.LinkByName(name)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		w.out.Error("VxLANWorker.UpVxLAN: %s", err)
	}
	if br, err := w.UpBr(cfg.Bridge); err != nil {
		return err
	} else if err := netlink.LinkSetMaster(link, br); err != nil {
		return err
	}
	return nil
}

func (w *VxLANWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	for _, mem := range w.inCfg.Members {
		if err := w.UpVxLAN(mem); err != nil {
			w.out.Error("VxLANWorker.Start %s %s", mem.Name, err)
		}
	}
}

func (w *VxLANWorker) DownVxLAN(cfg *config.VxLANMember) error {
	name := cfg.Name
	link, _ := netlink.LinkByName(name)
	if link == nil {
		return nil
	}
	port := &netlink.Vxlan{
		LinkAttrs: netlink.LinkAttrs{
			TxQLen: -1,
			Name:   name,
		},
	}
	if err := netlink.LinkDel(port); err != nil {
		return err
	}
	return nil
}

func (w *VxLANWorker) Stop() {
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

func (w *VxLANWorker) GetConfig() *config.Network {
	return w.cfg
}

func (w *VxLANWorker) GetSubnet() string {
	w.out.Warn("VxLANWorker.GetSubnet operation notSupport")
	return ""
}
