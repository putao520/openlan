package network

import "github.com/golang-collections/go-datastructures/queue"

type VirTap struct {
	isTap bool
	name  string
	inQ   *queue.Queue
	outQ  *queue.Queue
	bridge    Bridger
}

func NewVirTap(isTap bool, name string) (*VirTap, error) {
	tap := &VirTap{
		isTap: isTap,
		name:  name,
		inQ:   queue.New(1024*10),
		outQ:  queue.New(1024*10),
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
	result, err := t.inQ.Get(1)
	return copy(p, result[0].([]byte)), err
}

func (t *VirTap) Write(p []byte) (n int, err error) {
	return len(p), t.outQ.Put(p)
}

func (t *VirTap) Close() error {
	t.outQ.Dispose()
	t.inQ.Dispose()

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
