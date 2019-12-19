package point

import (
	"context"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/songgao/water"
	"net"
	"time"
)

type OnTapWorker interface {
	OnTap(*TapWorker) error
}

type TapWorker struct {
	writeChan chan []byte
	ifMtu     int
	doRecv    func([]byte) error
	config    *water.Config

	On     OnTapWorker
	Device *water.Interface
	//for tunnel device.
	EthDstAddr []byte
	EthSrcAddr []byte
	EthSrcIp   []byte
}

func NewTapWorker(devCfg *water.Config, c *config.Point, on OnTapWorker) (a *TapWorker) {
	a = &TapWorker{
		Device:    nil,
		config:    devCfg,
		writeChan: make(chan []byte, 1024*10),
		ifMtu:     c.IfMtu, //1514
		On:        on,
	}

	a.Open()
	if a.Device != nil && a.Device.IsTUN() {
		a.EthSrcIp = net.ParseIP(c.IfAddr).To4()
		libol.Info("NewTapWorker srcIp: % x", a.EthSrcIp)

		if c.IfEthSrc == "" {
			a.EthSrcAddr = libol.GenEthAddr(6)
		} else {
			if hw, err := net.ParseMAC(c.IfEthSrc); err == nil {
				a.EthSrcAddr = []byte(hw)
			}
		}
		if hw, err := net.ParseMAC(c.IfEthDst); err == nil {
			a.EthDstAddr = []byte(hw)
		}
		libol.Info("NewTapWorker src: % x, dst: % x", a.EthSrcAddr, a.EthDstAddr)
	}

	return
}

func (a *TapWorker) Open() {
	if a.Device != nil {
		a.Device.Close()
		time.Sleep(5 * time.Second) // sleep 5s and release cpu.
	}

	dev, err := water.New(*a.config)
	if err != nil {
		libol.Error("TapWorker.Open %s", err)
		return
	}
	libol.Info("TapWorker.Open %s", dev.Name())
	a.Device = dev
	if a.On != nil {
		a.On.OnTap(a)
	}
}

func (a *TapWorker) NewEth(t uint16) *libol.Ether {
	eth := libol.NewEther(t)
	eth.Dst = a.EthDstAddr
	eth.Src = a.EthSrcAddr

	return eth
}

func (a *TapWorker) GoRecv(ctx context.Context, doRecv func([]byte) error) {
	defer libol.Catch("TapWorker.GoRecv")
	defer a.Close()

	libol.Info("TapWorker.GoRev")
	a.doRecv = doRecv

	for {
		data := make([]byte, a.ifMtu)
		if a.Device == nil {
			return
		}

		n, err := a.Device.Read(data)
		if err != nil {
			libol.Error("TapWorker.GoRev: %s", err)
			a.Open()
			continue
		}

		libol.Debug("TapWorker.GoRev: % x", data[:n])
		if a.Device.IsTUN() {
			eth := a.NewEth(libol.ETHPIP4)

			buffer := make([]byte, 0, a.ifMtu)
			buffer = append(buffer, eth.Encode()...)
			buffer = append(buffer, data[0:n]...)
			n += eth.Len

			doRecv(buffer[:n])
		} else {
			doRecv(data[:n])
		}
	}
}

func (a *TapWorker) DoSend(data []byte) error {
	libol.Debug("TapWorker.DoSend: % x", data)

	a.writeChan <- data

	return nil
}

func (a *TapWorker) onArp(data []byte) bool {
	libol.Debug("TapWorker.onArp")
	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		libol.Warn("TapWorker.onArp %s", err)
		return false
	}

	if !eth.IsArp() {
		return false
	}

	arp, err := libol.NewArpFromFrame(data[eth.Len:])
	if err != nil {
		libol.Error("TapWorker.onArp %s.", err)
		return false
	}

	if arp.IsIP4() {
		if arp.OpCode != libol.ARP_REQUEST {
			return false
		}

		eth := a.NewEth(libol.ETHPARP)

		reply := libol.NewArp()
		reply.OpCode = libol.ARP_REPLY
		reply.SIpAddr = a.EthSrcIp
		reply.TIpAddr = arp.SIpAddr
		reply.SHwAddr = a.EthSrcAddr
		reply.THwAddr = arp.SHwAddr

		buffer := make([]byte, 0, a.ifMtu)
		buffer = append(buffer, eth.Encode()...)
		buffer = append(buffer, reply.Encode()...)

		libol.Info("TapWorker.onArp % x.", buffer)
		if a.doRecv != nil {
			a.doRecv(buffer)
		}

		return true
	}

	return false
}

func (a *TapWorker) GoLoop(ctx context.Context) {
	defer libol.Catch("TapWorker.GoLoop")
	defer a.Close()

	libol.Info("TapWorker.GoLoop")

	for {
		select {
		case w := <-a.writeChan:
			if a.Device == nil {
				return
			}

			if a.Device.IsTUN() {
				//Proxy arp request.
				if a.onArp(w) {
					libol.Info("TapWorker.GoLoop: Arp proxy.")
					continue
				}

				eth, err := libol.NewEtherFromFrame(w)
				if err != nil {
					libol.Error("TapWorker.GoLoop: %s", err)
					continue
				}
				if eth.IsVlan() {
					w = w[18:]
				} else if eth.IsIP4() {
					w = w[14:]
				} else { // default is Ethernet is 14 bytes.
					w = w[14:]
				}
			}

			if _, err := a.Device.Write(w); err != nil {
				libol.Error("TapWorker.GoLoop: %s", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (a *TapWorker) Close() {
	libol.Info("TapWorker.Close")

	if a.Device != nil {
		a.Device.Close()
		a.Device = nil
	}
}

func (a *TapWorker) Stop() {
	a.Close()
}
