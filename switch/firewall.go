package _switch

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/moby/libnetwork/iptables"
)

type FireWall struct {
	lock  libol.Locker
	rules []libol.FilterRule
}

func (f *FireWall) install() {
	for _, rule := range f.rules {
		if ret, err := libol.FilterRuleCmd(rule, "-I"); err != nil {
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
		if ret, err := libol.FilterRuleCmd(rule, "-D"); err != nil {
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
