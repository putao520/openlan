package olsw

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/moby/libnetwork/iptables"
	"sync"
)

type FireWall struct {
	lock  sync.Mutex
	rules []libol.IpTableRule
}

func (f *FireWall) install() {
	for _, rule := range f.rules {
		if ret, err := libol.IpTableCmd(rule, "-I"); err != nil {
			libol.Warn("FireWall.install %s", ret)
		}
	}
}

func (f *FireWall) Start() {
	f.lock.Lock()
	defer f.lock.Unlock()
	libol.Info("FireWall.Start")
	f.install()
	iptables.OnReloaded(func() {
		libol.Info("FireWall.Start OnReloaded")
		f.lock.Lock()
		defer f.lock.Unlock()
		f.install()
	})
}

func (f *FireWall) uninstall() {
	for _, rule := range f.rules {
		if ret, err := libol.IpTableCmd(rule, "-D"); err != nil {
			libol.Warn("FireWall.uninstall %s", ret)
		}
	}
}

func (f *FireWall) Stop() {
	f.lock.Lock()
	defer f.lock.Unlock()
	libol.Info("FireWall.Stop")
	f.uninstall()
}

func init() {
	libol.IpTableInit()
}
