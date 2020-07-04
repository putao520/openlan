package libol

import (
	"github.com/moby/libnetwork/iptables"
	"runtime"
)

type FilterRule struct {
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

func (rule FilterRule) Args() []string {
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
	}
	if rule.ToSource != "" {
		args = append(args, "--to-source", rule.ToSource)
	}
	if rule.ToDest != "" {
		args = append(args, "--to-destination", rule.ToDest)
	}
	return args
}

func FilterRuleCmd(rule FilterRule, opr string) ([]byte, error) {
	Debug("FilterRuleCmd: %s, %v", opr, rule)
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
		return nil, NewErr("iptables notSupport %s", runtime.GOOS)
	}
}

func init() {
	if err := iptables.FirewalldInit(); err != nil {
		Error("iptables.init %s", err)
	}
}
