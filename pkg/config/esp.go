package config

import (
	"fmt"
	"github.com/danieldin95/openlan/pkg/libol"
	"net"
	"strings"
)

const (
	EspAuth  = "8bc736635c0642aebc20ba5420c3e93a"
	EspCrypt = "4ac161f6635843b8b02c60cc36822515"
)

type EspState struct {
	Local    string `json:"local,omitempty" yaml:"local,omitempty"`
	LocalIp  net.IP `json:"-"  yaml:"-"`
	Remote   string `json:"remote,omitempty" yaml:"remote,omitempty"`
	RemoteIp net.IP `json:"-"  yaml:"-"`
	Encap    string `json:"encap,omitempty" yaml:"encapsulation,omitempty"`
	Auth     string `json:"auth,omitempty" yaml:"auth,omitempty"`
	Crypt    string `json:"crypt,omitempty" yaml:"crypt,omitempty"`
}

func (s *EspState) Pad32(value string) string {
	return strings.Repeat(value, 64/len(value))[:32]
}

func (s *EspState) Correct(obj *EspState) {
	if obj != nil {
		if s.Local == "" {
			s.Local = obj.Local
		}
		if s.Auth == "" {
			s.Auth = obj.Auth
		}
		if s.Crypt == "" {
			s.Crypt = obj.Crypt
		}
	}
	libol.Info("EspState.Correct: %s %s", s.Local, s.Remote)
	if s.Local == "" && s.Remote != "" {
		addr, _ := libol.GetLocalByGw(s.Remote)
		s.Local = addr.String()
	}
	if s.Crypt == "" {
		s.Crypt = s.Auth
	}
	if s.Auth == "" {
		s.Auth = EspAuth
	}
	if s.Crypt == "" {
		s.Crypt = EspCrypt
	}
	s.Auth = s.Pad32(s.Auth)
	s.Crypt = s.Pad32(s.Crypt)
}

type ESPPolicy struct {
	Source string `json:"source,omitempty"`
	Dest   string `json:"destination,omitempty"`
}

type ESPMember struct {
	Name     string       `json:"name"`
	Address  string       `json:"address,omitempty"`
	Peer     string       `json:"peer"`
	Spi      int          `json:"spi"`
	State    EspState     `json:"state"`
	Policies []*ESPPolicy `json:"policies" yaml:"policies,omitempty"`
}

type ESPSpecifies struct {
	Name    string       `json:"name"`
	Address string       `json:"address,omitempty"`
	State   EspState     `json:"state" yaml:"state,omitempty"`
	Members []*ESPMember `json:"members"`
}

func (n *ESPSpecifies) Correct() {
	for _, m := range n.Members {
		if m.Address == "" {
			m.Address = n.Address
		}
		if m.Address == "" {
			libol.Warn("ESPSpecifies.Correct %s need address", n.Name)
			continue
		}
		if !strings.Contains(m.Address, "/") {
			m.Address += "/32"
		}
		if !strings.Contains(m.Peer, "/") {
			m.Peer += "/32"
		}
		ptr := &m.State
		ptr.Correct(&n.State)
		if m.Policies == nil {
			m.Policies = make([]*ESPPolicy, 0, 2)
		}
		m.Policies = append(m.Policies, &ESPPolicy{
			Source: m.Address,
			Dest:   m.Peer,
		})
		if m.Spi == 0 {
			m.Spi = libol.GenInt32()
		}
		if m.Name == "" {
			m.Name = fmt.Sprintf("spi%d", m.Spi)
		}
	}
}
