package config

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"net"
	"strings"
)

const (
	EspAuth  = "8bc736635c0642aebc20ba5420c3e93a"
	EspCrypt = "4ac161f6635843b8b02c60cc36822515"
)

type EspState struct {
	Local    string `json:"local,omitempty"`
	LocalIp  net.IP `json:"-"`
	Remote   string `json:"remote"`
	RemoteIp net.IP `json:"-"`
	Auth     string `json:"auth,omitempty"`
	Crypt    string `json:"crypt,omitempty"`
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
	Spi      uint32       `json:"spi"`
	State    EspState     `json:"state"`
	Policies []*ESPPolicy `json:"policies"`
}

type ESPInterface struct {
	Mode    string       `json:"mode"`
	Name    string       `json:"name"`
	Address string       `json:"address,omitempty"`
	State   EspState     `json:"state"`
	Members []*ESPMember `json:"members"`
}

func (n *ESPInterface) Correct() {
	for _, m := range n.Members {
		if m.Address == "" {
			m.Address = n.Address
		}
		if m.Address == "" {
			libol.Warn("ESPInterface.Correct %s need address", n.Name)
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
			m.Spi = libol.GenUint32()
		}
		if m.Name == "" {
			m.Name = fmt.Sprintf("spi%d", m.Spi)
		}
	}
}
