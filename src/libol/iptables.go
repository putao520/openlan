package libol

import (
	"github.com/moby/libnetwork/iptables"
	"runtime"
)

type IptChain struct {
	Table string
	Name  string
}

type IptRule struct {
	Table    string
	Chain    string
	Input    string
	Source   string
	ToSource string
	Dest     string
	ToDest   string
	Output   string
	Comment  string
	Jump     string
}

func (rule IptRule) Args() []string {
	var args []string
	if rule.Source != "" {
		args = append(args, "-s", rule.Source)
	}
	if rule.Input != "" {
		args = append(args, "-i", rule.Input)
	}
	if rule.Dest != "" {
		args = append(args, "-d", rule.Dest)
	}
	if rule.Output != "" {
		args = append(args, "-o", rule.Output)
	}
	if rule.Jump != "" {
		args = append(args, "-j", rule.Jump)
	} else {
		args = append(args, "-j", "ACCEPT")
	}
	if rule.ToSource != "" {
		args = append(args, "--to-source", rule.ToSource)
	}
	if rule.ToDest != "" {
		args = append(args, "--to-destination", rule.ToDest)
	}
	return args
}

func IptRCmd(rule IptRule, opr string) ([]byte, error) {
	Debug("IptRCmd: %s, %v", opr, rule)
	table := iptables.Table(rule.Table)
	chain := rule.Chain
	switch runtime.GOOS {
	case "linux":
		args := rule.Args()
		fullArgs := append([]string{"-t", rule.Table, opr, rule.Chain}, args...)
		if opr == "-I" || opr == "-A" {
			if iptables.Exists(table, chain, args...) {
				return nil, nil
			}
		}
		return iptables.Raw(fullArgs...)
	default:
		return nil, NewErr("IpTable notSupport %s", runtime.GOOS)
	}
}

func IptCCmd(chain IptChain, opr string) ([]byte, error) {
	Debug("IptCCmd: %s, %v", opr, chain)
	table := iptables.Table(chain.Table)
	name := chain.Name
	switch runtime.GOOS {
	case "linux":
		if opr == "-N" {
			if iptables.ExistChain(name, table) {
				return nil, nil
			}
			if _, err := iptables.NewChain(name, table, true); err != nil {
				return nil, err
			}
		} else if opr == "-X" { // -X
			if err := iptables.RemoveExistingChain(name, table); err != nil {
				return nil, err
			}
		}
	default:
		return nil, NewErr("IpTable notSupport %s", runtime.GOOS)
	}
	return nil, nil
}

func IptInit() {
	if err := iptables.FirewalldInit(); err != nil {
		Error("IpTables.init %s", err)
	}
}
