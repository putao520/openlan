package point

import (
	"encoding/binary"
	"github.com/danieldin95/openlan-go/libol"
	"sync"
	"time"
)

type Neighbor struct {
	HwAddr  []byte
	IpAddr  []byte
	Uptime  int64
	NewTime int64
}

type Neighbors struct {
	lock      sync.RWMutex
	neighbors map[uint32]*Neighbor
	done      chan bool
	ticker    *time.Ticker
	timeout   int64
}

func (n *Neighbors) Expire() {
	deletes := make([]uint32, 0, 1024)

	n.lock.Lock()
	defer n.lock.Unlock()
	//collect need deleted.
	for index, learn := range n.neighbors {
		now := time.Now().Unix()
		if now-learn.Uptime > n.timeout {
			deletes = append(deletes, index)
		}
	}
	libol.Debug("Neighbors.Expire delete %d", len(deletes))
	//execute delete.
	for _, d := range deletes {
		if l, ok := n.neighbors[d]; ok {
			delete(n.neighbors, d)
			libol.Info("Neighbors.Expire: delete %x", l.HwAddr)
		}
	}
}

func (n *Neighbors) Start() {
	for {
		select {
		case <-n.done:
			return
		case t := <-n.ticker.C:
			libol.Log("Neighbors.Expire: tick at %s", t)
			n.Expire()
		}
	}
}

func (n *Neighbors) Stop() {
	n.ticker.Stop()
	n.done <- true
}

func (n *Neighbors) Add(h *Neighbor) {
	if h == nil {
		return
	}
	n.lock.Lock()
	defer n.lock.Unlock()
	k := binary.BigEndian.Uint32(h.IpAddr)
	if l, ok := n.neighbors[k]; ok {
		l.Uptime = h.Uptime
		copy(l.HwAddr[:6], h.HwAddr[:6])
	} else {
		l := &Neighbor{
			Uptime:  h.Uptime,
			NewTime: h.NewTime,
			HwAddr:  make([]byte, 6),
			IpAddr:  make([]byte, 4),
		}
		copy(l.IpAddr[:4], h.IpAddr[:4])
		copy(l.HwAddr[:6], h.HwAddr[:6])
		n.neighbors[k] = l
	}
}

func (n *Neighbors) Get(d uint32) *Neighbor {
	n.lock.RLock()
	defer n.lock.RUnlock()
	if l, ok := n.neighbors[d]; ok {
		return l
	}
	return nil
}

func (n *Neighbors) Clear() {
	libol.Info("Neighbor.Clear")
	n.lock.Lock()
	defer n.lock.Unlock()

	deletes := make([]uint32, 0, 1024)
	for index := range n.neighbors {
		deletes = append(deletes, index)
	}
	//execute delete.
	for _, d := range deletes {
		if _, ok := n.neighbors[d]; ok {
			delete(n.neighbors, d)
		}
	}
}

func (n *Neighbors) GetByBytes(d []byte) *Neighbor {
	n.lock.RLock()
	defer n.lock.RUnlock()

	k := binary.BigEndian.Uint32(d)
	if l, ok := n.neighbors[k]; ok {
		return l
	}
	return nil
}
