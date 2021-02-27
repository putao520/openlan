package olsw

import (
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/moby/libnetwork/iptables"
	"sync"
)

const (
	NatT         = "nat"
	FilterT      = "filter"
	InputC       = "INPUT"
	ForwardC     = "FORWARD"
	OutputC      = "OUTPUT"
	PostRoutingC = "POSTROUTING"
	PreRoutingC  = "PREROUTING"
	MasqueradeC  = "MASQUERADE"
	OlInputC     = "openlan_IN"
	OlForwardC   = "openlan_FWD"
	OlOutputC    = "openlan_OUT"
	OlPreC       = "openlan_PRE"
	OlPostC      = "openlan_POST"
)

type FireWall struct {
	lock   sync.Mutex
	chains network.IpChains
	rules  network.IpRules
}

func NewFireWall(flows []config.FlowRule) *FireWall {
	f := &FireWall{
		chains: make(network.IpChains, 0, 8),
		rules:  make(network.IpRules, 0, 32),
	}
	// Load custom rules.
	for _, rule := range flows {
		f.rules = f.rules.Add(network.IpRule{
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
	f.AddChain(network.IpChain{
		Table: FilterT,
		Name:  OlInputC,
	})
	f.AddChain(network.IpChain{
		Table: FilterT,
		Name:  OlForwardC,
	})
	f.AddChain(network.IpChain{
		Table: FilterT,
		Name:  OlOutputC,
	})
	f.AddChain(network.IpChain{
		Table: NatT,
		Name:  OlPostC,
	})
	// Enable chains
	f.AddRule(network.IpRule{
		Table: FilterT,
		Chain: InputC,
		Jump:  OlInputC,
	})
	f.AddRule(network.IpRule{
		Table: FilterT,
		Chain: ForwardC,
		Jump:  OlForwardC,
	})
	f.AddRule(network.IpRule{
		Table: FilterT,
		Chain: OutputC,
		Jump:  OlOutputC,
	})
	f.AddRule(network.IpRule{
		Table: NatT,
		Chain: PostRoutingC,
		Jump:  OlPostC,
	})
	libol.Info("FireWall.Initialize total %d rules", len(f.rules))
}

func (f *FireWall) AddChain(chain network.IpChain) {
	f.chains = f.chains.Add(chain)
}

func (f *FireWall) AddRule(rule network.IpRule) {
	f.rules = f.rules.Add(rule)
}

func (f *FireWall) ApplyRule(rule network.IpRule) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	if _, err := rule.Opr("-I"); err != nil {
		return err
	}
	f.rules = f.rules.Add(rule)
	return nil
}

func (f *FireWall) install() {
	for _, c := range f.chains {
		if _, err := c.Opr("-N"); err != nil {
			libol.Warn("FireWall.install %s", err)
		}
	}
	for _, r := range f.rules {
		if ret, err := r.Opr("-I"); err != nil {
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

func (f *FireWall) RevokeRule(rule network.IpRule) error {
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
	network.IpInit()
}
