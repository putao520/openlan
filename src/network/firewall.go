package network

import (
	"github.com/danieldin95/openlan/src/config"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/moby/libnetwork/iptables"
	"sync"
)

const (
	OLCInput   = "INPUT_direct"
	OLCForward = "FORWARD_direct"
	OLCOutput  = "OUTPUT_direct"
	OLCPre     = "PREROUTING_direct"
	OLCPost    = "POSTROUTING_direct"
)

type FireWall struct {
	lock   sync.Mutex
	chains IpChains
	rules  IpRules
}

func NewFireWall(flows []config.FlowRule) *FireWall {
	f := &FireWall{
		chains: make(IpChains, 0, 8),
		rules:  make(IpRules, 0, 32),
	}
	// Load custom rules.
	for _, rule := range flows {
		f.rules = f.rules.Add(IpRule{
			Table:    rule.Table,
			Chain:    rule.Chain,
			Source:   rule.Source,
			Dest:     rule.Dest,
			Jump:     rule.Jump,
			ToSource: rule.ToSource,
			ToDest:   rule.ToDest,
			Comment:  rule.Comment,
			Input:    rule.Input,
			Output:   rule.Output,
		})
	}
	return f
}

func (f *FireWall) Initialize() {
	// Init chains
	f.AddChain(IpChain{
		Table: TFilter,
		Name:  OLCInput,
	})
	f.AddChain(IpChain{
		Table: TFilter,
		Name:  OLCForward,
	})
	f.AddChain(IpChain{
		Table: TFilter,
		Name:  OLCOutput,
	})
	f.AddChain(IpChain{
		Table: TNat,
		Name:  OLCPost,
	})
	libol.Info("FireWall.Initialize %d chains", len(f.chains))
	// Enable chains
	f.AddRule(IpRule{
		Order: "-I",
		Table: TFilter,
		Chain: CInput,
		Jump:  OLCInput,
	})
	f.AddRule(IpRule{
		Order: "-I",
		Table: TFilter,
		Chain: CForward,
		Jump:  OLCForward,
	})
	f.AddRule(IpRule{
		Order: "-I",
		Table: TFilter,
		Chain: COutput,
		Jump:  OLCOutput,
	})
	f.AddRule(IpRule{
		Order: "-I",
		Table: TNat,
		Chain: CPostRoute,
		Jump:  OLCPost,
	})
	libol.Info("FireWall.Initialize %d rules", len(f.rules))
}

func (f *FireWall) AddChain(chain IpChain) {
	f.chains = f.chains.Add(chain)
}

func (f *FireWall) AddRule(rule IpRule) {
	f.rules = f.rules.Add(rule)
}

func (f *FireWall) ApplyRule(rule IpRule) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	order := rule.Order
	if order == "" {
		order = "-A"
	}
	if _, err := rule.Opr(order); err != nil {
		return err
	}
	f.rules = f.rules.Add(rule)
	return nil
}

func (f *FireWall) install() {
	for _, c := range f.chains {
		if _, err := c.Opr("-N"); err != nil {
			libol.Error("FireWall.install %s", err)
		}
	}
	for _, r := range f.rules {
		order := r.Order
		if order == "" {
			order = "-A"
		}
		if _, err := r.Opr(order); err != nil {
			libol.Error("FireWall.install %s", err)
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
		if ret, err := rule.Opr("-D"); err != nil {
			libol.Warn("FireWall.uninstall %s", ret)
		}
	}
	for _, c := range f.chains {
		if _, err := c.Opr("-X"); err != nil {
			libol.Warn("FireWall.uninstall %s", err)
		}
	}
}

func (f *FireWall) RevokeRule(rule IpRule) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	if _, err := rule.Opr("-D"); err != nil {
		return err
	}
	f.rules = f.rules.Pop(rule)
	return nil
}

func (f *FireWall) Stop() {
	f.lock.Lock()
	defer f.lock.Unlock()
	libol.Info("FireWall.Stop")
	f.uninstall()
}

func (f *FireWall) Refresh() {
	f.uninstall()
	f.install()
}

func init() {
	IpInit()
}
