package _switch

import "github.com/danieldin95/openlan-go/libol"

type FireWall struct {
	Rules []libol.FilterRule
}

func (f *FireWall) Start() {
	for _, rule := range f.Rules {
		if ret, err := libol.IPTables(rule, "-I"); err != nil {
			libol.Warn("FireWall.Start %s", ret)
		}
	}
}

func (f *FireWall) Stop() {
	for _, rule := range f.Rules {
		if ret, err := libol.IPTables(rule, "-D"); err != nil {
			libol.Warn("FireWall.Start %s", ret)
		}
	}
}
