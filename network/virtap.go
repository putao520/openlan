package network

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/golang-collections/go-datastructures/queue"
)

type VirTap struct {
	isTap  bool
	name   string
	writeQ *queue.Queue
	readQ  *queue.Queue
	bridge Bridger
}

func NewVirTap(isTap bool, name string) (*VirTap, error) {
	if name == "" {
		name = Tapers.GenName()
	}
	tap := &VirTap{
		isTap:  isTap,
		name:   name,
		writeQ: queue.New(1024 * 10),
		readQ:  queue.New(1024 * 10),
	}
	Tapers.Add(tap)

	return tap, nil
}

func (t *VirTap) IsTUN() bool {
	return !t.isTap
}

func (t *VirTap) IsTAP() bool {
	return t.isTap
}

func (t *VirTap) Name() string {
	return t.name
}

func (t *VirTap) Read(p []byte) (n int, err error) {
	result, err := t.readQ.Get(1)
	if err == nil {
		return copy(p, result[0].([]byte)), err
	}
	return 0, err
}

func (t *VirTap) InRead(p []byte) (n int, err error) {
	libol.Debug("VirTap.InRead: %s % x", t, p[:20])
	return len(p), t.readQ.Put(p)
}

func (t *VirTap) Write(p []byte) (n int, err error) {
	libol.Debug("VirTap.Write: %s % x", t, p[:20])
	return len(p), t.writeQ.Put(p)
}

func (t *VirTap) OutWrite() ([]byte, error) {
	result, err := t.writeQ.Get(1)
	if err != nil {
		return nil, err
	}
	return result[0].([]byte), nil
}

func (t *VirTap) Deliver() {
	for {
		data, err := t.OutWrite()
		if err != nil {
			break
		}
		libol.Debug("VirTap.Deliver: %s % x", t, data[:20])
		if t.bridge == nil {
			continue
		}

		m := &Framer{Data: data, Source: t}
		t.bridge.Input(m)
	}
}

func (t *VirTap) Close() error {
	t.readQ.Dispose()

	if t.bridge != nil {
		t.bridge.DelSlave(t)
		t.bridge = nil
	}
	Tapers.Del(t.name)
	return nil
}

func (t *VirTap) Slave(bridge Bridger) {
	if t.bridge == nil {
		t.bridge = bridge
	}
}

func (t *VirTap) Up() {
	go t.Deliver()
}

func (t *VirTap) String() string {
	return t.name
}
