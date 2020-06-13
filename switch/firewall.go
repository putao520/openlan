package _switch

import "github.com/danieldin95/openlan-go/libol"

type FireWall struct {
	lock  libol.Locker
	rules []libol.FilterRule
}

func (f *FireWall) Start() {
	f.lock.Lock()
	defer f.lock.Unlock()
	for _, rule := range f.rules {
		if ret, err := libol.IPTables(rule, "-I"); err != nil {
			libol.Warn("FireWall.Start %s", ret)
		}
	}
}

func (f *FireWall) Stop() {
	f.lock.Lock()
	defer f.lock.Unlock()
	for _, rule := range f.rules {
		if ret, err := libol.IPTables(rule, "-D"); err != nil {
			libol.Warn("FireWall.Start %s", ret)
		}
	}
}
