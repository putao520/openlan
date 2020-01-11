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

type TapWorkerListener struct {
	OnOpen  func(w *TapWorker) error
	OnClose func(w *TapWorker)
	ReadAt  func([]byte) error
}

type TapWorker struct {
	Device   network.Taper
	Listener TapWorkerListener
	//for tunnel device.
	EthDstAddr []byte
	EthSrcAddr []byte
	EthSrcIp   []byte

	writeChan chan []byte
	devCfg   *water.Config
	pointCfg *config.Point
}

func NewTapWorker(devCfg *water.Config, c *config.Point) (a *TapWorker) {
	a = &TapWorker{
		Device:    nil,
		devCfg:    devCfg,
		pointCfg:  c,
		writeChan: make(chan []byte, 1024*10),
	}

	return
}

func (a *TapWorker) DoTun() {
	if a.Device == nil || !a.Device.IsTun() {
		return
	}

	a.EthSrcIp = net.ParseIP(a.pointCfg.IfAddr).To4()
	libol.Info("NewTapWorker srcIp: % x", a.EthSrcIp)

	if a.pointCfg.IfEthSrc == "" {
		a.EthSrcAddr = libol.GenEthAddr(6)
	} else {
		if hw, err := net.ParseMAC(a.pointCfg.IfEthSrc); err == nil {
			a.EthSrcAddr = []byte(hw)
		}
	}
	if hw, err := net.ParseMAC(a.pointCfg.IfEthDst); err == nil {
		a.EthDstAddr = []byte(hw)
	}
	libol.Info("NewTapWorker src: %x, dst: %x", a.EthSrcAddr, a.EthDstAddr)

}
func (a *TapWorker) Start(ctx context.Context, p Pointer) {
	a.Open()
	a.DoTun()

	go a.Read(ctx)
	go a.Loop(ctx)
}
func (a *TapWorker) Open() {
	if a.Device != nil {
		a.Device.Close()
		time.Sleep(5 * time.Second) // sleep 5s and release cpu.
	}

	var err error
	var dev network.Taper
	if a.devCfg.DeviceType == water.TAP {
		dev, err = network.NewLinuxTap(true, "")
	} else {
		dev, err = network.NewLinuxTap(false, "")
	}

	if err != nil {
		libol.Error("TapWorker.Open %s", err)
		return
	}

	libol.Info("TapWorker.Open %s", dev.Name())
	a.Device = dev
	if a.Listener.OnOpen != nil {
		a.Listener.OnOpen(a)
	}
}

func (a *TapWorker) NewEth(t uint16) *libol.Ether {
	eth := libol.NewEther(t)
	eth.Dst = a.EthDstAddr
	eth.Src = a.EthSrcAddr

	return eth
}

func (a *TapWorker) Read(ctx context.Context) {
	defer libol.Catch("TapWorker.Read")
	defer a.Close()

	libol.Info("TapWorker.Read")
	for {
		if a.Device == nil {
			return
		}

		data := make([]byte, a.pointCfg.IfMtu)
		n, err := a.Device.Read(data)
		if err != nil {
			libol.Error("TapWorker.Read: %s", err)
			a.Open()
			continue
		}

		libol.Debug("TapWorker.Read: % x", data[:n])
		if a.Device.IsTun() {
			eth := a.NewEth(libol.ETHPIP4)

			buffer := make([]byte, 0, a.pointCfg.IfMtu)
			buffer = append(buffer, eth.Encode()...)
			buffer = append(buffer, data[0:n]...)
			n += eth.Len
			if a.Listener.ReadAt != nil {
				a.Listener.ReadAt(buffer[:n])
			}
		} else {
			if a.Listener.ReadAt != nil {
				a.Listener.ReadAt(data[:n])
			}
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

		buffer := make([]byte, 0, a.pointCfg.IfMtu)
		buffer = append(buffer, eth.Encode()...)
		buffer = append(buffer, reply.Encode()...)

		libol.Info("TapWorker.onArp % x.", buffer)
		if a.Listener.ReadAt != nil {
			a.Listener.ReadAt(buffer)
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
		if a.Listener.OnClose != nil {
			a.Listener.OnClose(a)
		}
		a.Device.Close()
		a.Device = nil
	}
}

func (a *TapWorker) Stop() {
	close(a.writeChan)
	a.Close()
	a.Device = nil
}
