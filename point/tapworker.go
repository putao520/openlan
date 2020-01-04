package point

import (
	"context"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/network"
	"github.com/songgao/water"
	"net"
	"time"
)

type OnTapWorker interface {
	OnTap(w *TapWorker) error
}

type TapWorker struct {
	writeChan chan []byte
	ifMtu     int
	doRead    func([]byte) error
	config    *water.Config

	On     OnTapWorker
	Device network.Taper
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
	if a.Device != nil && a.Device.IsTun() {
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

	var err error
	var dev network.Taper
	if a.config.DeviceType == water.TAP {
		dev, err = network.NewLinTap(true, "")
	} else {
		dev, err = network.NewLinTap(false, "")
	}

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

func (a *TapWorker) Read(ctx context.Context, doRead func(p []byte) error) {
	defer libol.Catch("TapWorker.Read")
	defer a.Close()

	libol.Info("TapWorker.Read")
	a.doRead = doRead
	data := make([]byte, a.ifMtu)

	for {
		if a.Device == nil {
			return
		}

		n, err := a.Device.Read(data)
		if err != nil {
			libol.Error("TapWorker.Read: %s", err)
			a.Open()
			continue
		}

		libol.Debug("TapWorker.Read: % x", data[:n])
		if a.Device.IsTun() {
			eth := a.NewEth(libol.ETHPIP4)

			buffer := make([]byte, 0, a.ifMtu)
			buffer = append(buffer, eth.Encode()...)
			buffer = append(buffer, data[0:n]...)
			n += eth.Len

			doRead(buffer[:n])
		} else {
			doRead(data[:n])
		}
	}
}

func (a *TapWorker) DoWrite(data []byte) error {
	libol.Debug("TapWorker.DoWrite: % x", data)

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
		if a.doRead != nil {
			a.doRead(buffer)
		}

		return true
	}

	return false
}

func (a *TapWorker) Loop(ctx context.Context) {
	defer libol.Catch("TapWorker.Loop")
	defer a.Close()

	libol.Info("TapWorker.Loop")

	for {
		select {
		case w := <-a.writeChan:
			if a.Device == nil {
				return
			}

			if a.Device.IsTun() {
				//Proxy arp request.
				if a.onArp(w) {
					libol.Info("TapWorker.Loop: Arp proxy.")
					continue
				}

				eth, err := libol.NewEtherFromFrame(w)
				if err != nil {
					libol.Error("TapWorker.Loop: %s", err)
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
				libol.Error("TapWorker.Loop: %s", err)
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
