package libol

import (
	"fmt"
	"os/exec"
	"runtime"
)

type IpFilterRule struct {
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

func IPTables(rule IpFilterRule, action string) (string, error) {
	switch runtime.GOOS {
	case "linux":
		args := []string{"-t", rule.Table, action, rule.Chain}
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
		ret, err := exec.Command("/usr/sbin/iptables", args...).CombinedOutput()
		return fmt.Sprintf("%v: %s", args, ret), err
	default:
		return "", NewErr("iptables %s not support", runtime.GOOS)
	}
}
