package config

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lightstar-dev/openlan-go/libol"
)

type Script struct {
	Cmd  string
	Data interface{}
}

func NewScript(cmd string) *Script {
	path, err := filepath.Abs(cmd)
	if err != nil {
		path = cmd
	}
	s := Script{
		Cmd:  path,
		Data: nil,
	}
	return &s
}

func (s *Script) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

func (s *Script) CallBefore() {
	if _, err := os.Stat(s.Cmd); os.IsNotExist(err) {
		libol.Warn("Script.RunBefore:<%s> does not exist", s.Cmd)
		return
	}

	out, err := s.Run(s.Cmd, "before")
	if err == nil {
		libol.Info("%s before: %s", s.Cmd, string(out))
	} else {
		libol.Warn("%s before: %s", s.Cmd, err)
	}
}

func (s *Script) CallAfter() {
	if _, err := os.Stat(s.Cmd); os.IsNotExist(err) {
		libol.Warn("Script.RunAfter:<%s> does not exist", s.Cmd)
		return
	}

	out, err := s.Run(s.Cmd, "after")
	if err == nil {
		libol.Info("%s after: %s", s.Cmd, string(out))
	} else {
		libol.Warn("%s after: %s", s.Cmd, err)
	}
}

func (s *Script) CallExit() {
	if _, err := os.Stat(s.Cmd); os.IsNotExist(err) {
		libol.Warn("Script.RunAfter:<%s> does not exist", s.Cmd)
		return
	}

	out, err := s.Run(s.Cmd, "exit")
	if err == nil {
		libol.Info("%s exit: %s", s.Cmd, string(out))
	} else {
		libol.Warn("%s exit: %s", s.Cmd, err)
	}
}

