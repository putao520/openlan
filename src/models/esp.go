// +build linux

package models

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/schema"
	nl "github.com/vishvananda/netlink"
	"net"
)

type Esp struct {
	Name    string
	Address string
}

func (l *Esp) Update() {
}

func (l *Esp) ID() string {
	return l.Name
}

func NewEspSchema(e *Esp) schema.Esp {
	e.Update()
	se := schema.Esp{
		Name:    e.Name,
		Address: e.Address,
	}
	return se
}

type EspState struct {
	*schema.EspState
}

func (l *EspState) Update() {
	xs := &nl.XfrmState{
		Spi:   l.Spi,
		Src:   net.ParseIP(l.Source),
		Dst:   net.ParseIP(l.Dest),
		Mode:  nl.Mode(l.Mode),
		Proto: nl.Proto(l.Proto),
	}
	if xss, err := nl.XfrmStateGet(xs); xss != nil {
		l.TxBytes = xss.Statistics.Bytes
		l.TxPackages = xss.Statistics.Packets
	} else {
		libol.Debug("EspState.Update %s", err)
	}
	xs.Src, xs.Dst = xs.Dst, xs.Src
	if xss, err := nl.XfrmStateGet(xs); xss != nil {
		l.RxBytes = xss.Statistics.Bytes
		l.RxPackages = xss.Statistics.Packets
	} else {
		libol.Debug("EspState.Update %s", err)
	}
}

func (l *EspState) ID() string {
	return fmt.Sprintf("%d-%s-%s", l.Spi, l.Source, l.Dest)
}

func NewEspStateSchema(e *EspState) schema.EspState {
	e.Update()
	se := schema.EspState{
		Name:       e.Name,
		Spi:        e.Spi,
		Source:     e.Source,
		Dest:       e.Dest,
		TxBytes:    e.TxBytes,
		TxPackages: e.TxPackages,
		RxBytes:    e.RxBytes,
		RxPackages: e.RxPackages,
	}
	return se
}

type EspPolicy struct {
	*schema.EspPolicy
}

func (l *EspPolicy) Update() {
}

func (l *EspPolicy) ID() string {
	return fmt.Sprintf("%d-%s-%s", l.Spi, l.Source, l.Dest)
}

func NewEspPolicySchema(e *EspPolicy) schema.EspPolicy {
	e.Update()
	se := schema.EspPolicy{
		Name:   e.Name,
		Source: e.Source,
		Dest:   e.Dest,
	}
	return se
}
