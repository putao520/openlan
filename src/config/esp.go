package config

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"strings"
)

type EspState struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Auth   string `json:"auth"`
	Crypt  string `json:"crypt"`
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
	if s.Crypt == "" {
		s.Crypt = s.Auth
	}
}

type ESPPolicy struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type ESPMember struct {
	Name     string       `json:"name"`
	Address  string       `json:"address"`
	Peer     string       `json:"peer"`
	Spi      uint32       `json:"spi"`
	State    EspState     `json:"state"`
	Policies []*ESPPolicy `json:"policies"`
}

type ESPInterface struct {
	Name    string       `json:"name"`
	Address string       `json:"address"`
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
		s := &m.State
		s.Correct(&n.State)
		if m.Policies == nil {
			m.Policies = make([]*ESPPolicy, 0, 2)
		}
		m.Policies = append(m.Policies, &ESPPolicy{
			Source:      m.Address,
			Destination: m.Peer,
		})
		if m.Spi == 0 {
			m.Spi = libol.GenUint32()
		}
		if m.Name == "" {
			m.Name = fmt.Sprintf("esp-%d", m.Spi)
		}
	}
}
