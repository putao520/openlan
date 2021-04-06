package config

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
)

type EspState struct {
	Local   string `json:"local"`
	Private string `json:"private"`
	Remote  string `json:"remote"`
	Auth    string `json:"auth"`
	Crypt   string `json:"crypt"`
}

func (s *EspState) Correct(obj *EspState) {
	if s.Local == "" {
		s.Local = obj.Local
	}
	if s.Auth == "" {
		s.Auth = obj.Auth
	}
	if s.Crypt == "" {
		s.Crypt = obj.Crypt
	}
	if s.Private == "" {
		s.Private = obj.Private
	}
	if s.Private == "" {
		s.Private = s.Local
	}
}

type ESPPolicy struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type ESPMember struct {
	Name     string       `json:"name"`
	Local    string       `json:"local"`
	Remote   string       `json:"remote"`
	Spi      uint32       `json:"spi"`
	State    EspState     `json:"state"`
	Policies []*ESPPolicy `json:"policies"`
}

type ESPInterface struct {
	Name    string       `json:"name"`
	Local   string       `json:"local"`
	State   EspState     `json:"state"`
	Members []*ESPMember `json:"members"`
}

func (n *ESPInterface) Correct() {
	for _, m := range n.Members {
		if m.Local == "" {
			m.Local = n.Local
		}
		if m.Local == "" {
			libol.Warn("ESPInterface.Correct %s need local", n.Name)
			continue
		}
		s := &m.State
		s.Correct(&n.State)
		if m.Policies == nil {
			m.Policies = make([]*ESPPolicy, 0, 2)
		}
		m.Policies = append(m.Policies, &ESPPolicy{
			Source:      m.Local + "/32",
			Destination: m.Remote + "/32",
		})
		if m.Spi == 0 {
			m.Spi = libol.GenUint32()
		}
		if m.Name == "" {
			m.Name = fmt.Sprintf("esp-%d", m.Spi)
		}
	}
}
