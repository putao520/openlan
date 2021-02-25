package olsw

import (
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
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
	OlInputC     = "openlan-IN"
	OlForwardC   = "openlan-FWD"
	OlOutputC    = "openlan-OUT"
	OlPreC       = "openlan-PRE"
	OlPostC      = "openlan-POST"
)

type FireWall struct {
	lock   sync.Mutex
	chains []libol.IptChain
	rules  []libol.IptRule
}

func NewFireWall(flows []config.FlowRule) *FireWall {
	f := &FireWall{
		chains: make([]libol.IptChain, 0, 8),
		rules:  make([]libol.IptRule, 0, 32),
	}
	// Load custom rules.
	for _, rule := range flows {
		f.rules = append(f.rules, libol.IptRule{
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
	f.AddChain(libol.IptChain{
		Table: FilterT,
		Name:  OlInputC,
	})
	f.AddChain(libol.IptChain{
		Table: FilterT,
		Name:  OlForwardC,
	})
	f.AddChain(libol.IptChain{
		Table: FilterT,
		Name:  OlOutputC,
	})
	f.AddChain(libol.IptChain{
		Table: NatT,
		Name:  OlPostC,
	})
	// Enable chains
	f.AddRule(libol.IptRule{
		Table: FilterT,
		Chain: InputC,
		Jump:  OlInputC,
	})
	f.AddRule(libol.IptRule{
		Table: FilterT,
		Chain: ForwardC,
		Jump:  OlForwardC,
	})
	f.AddRule(libol.IptRule{
		Table: FilterT,
		Chain: OutputC,
		Jump:  OlOutputC,
	})
	f.AddRule(libol.IptRule{
		Table: NatT,
		Chain: PostRoutingC,
		Jump:  OlPostC,
	})
	libol.Info("FireWall.Initialize total %d rules", len(f.rules))
}

func (f *FireWall) AddChain(chain libol.IptChain) {
	f.chains = append(f.chains, chain)
}

func (f *FireWall) AddRule(rule libol.IptRule) {
	f.rules = append(f.rules, rule)
}

func (f *FireWall) install() {
	for _, c := range f.chains {
		if _, err := libol.IptChainOpr(c, "-N"); err != nil {
			libol.Warn("FireWall.install %s", err)
		}
	}
	for _, r := range f.rules {
		if ret, err := libol.IptRuleOpr(r, "-I"); err != nil {
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
		if ret, err := libol.IptRuleOpr(rule, "-D"); err != nil {
			libol.Warn("FireWall.uninstall %s", ret)
		}
	}
	for _, c := range f.chains {
		if _, err := libol.IptChainOpr(c, "-X"); err != nil {
			libol.Warn("FireWall.uninstall %s", err)
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
	libol.IptInit()
}
