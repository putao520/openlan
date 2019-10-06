package point

import (
	"net"

	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/songgao/water"
)

type TapWorker struct {
	writechan chan []byte
	ifmtu     int
	doRecv    func([]byte) error

	Ifce *water.Interface
	//for tunnel device.
	EthDstAddr []byte
	EthSrcAddr []byte
	EthSrcIp   []byte
}

func NewTapWorker(ifce *water.Interface, c *Config) (a *TapWorker) {
	a = &TapWorker{
		Ifce:      ifce,
		writechan: make(chan []byte, 1024*10),
		ifmtu:     c.Ifmtu, //1514
	}

	if a.Ifce.IsTUN() {
		a.EthSrcIp = net.ParseIP(c.Ifaddr).To4()
		libol.Info("NewTapWoker srcIp: % x", a.EthSrcIp)

		if c.Ifethsrc == "" {
			a.EthSrcAddr = libol.GenEthAddr(6)
		} else {
			if hw, err := net.ParseMAC(c.Ifethsrc); err == nil {
				a.EthSrcAddr = []byte(hw)
			}
		}
		if hw, err := net.ParseMAC(c.Ifethdst); err == nil {
			a.EthDstAddr = []byte(hw)
		}
		libol.Info("NewTapWorker src: % x, dst: % x", a.EthSrcAddr, a.EthDstAddr)
	}

	return
}

func (a *TapWorker) NewEth(t uint16) *libol.Ether {
	eth := libol.NewEther(t)
	eth.Dst = a.EthDstAddr
	eth.Src = a.EthSrcAddr

	return eth
}

func (a *TapWorker) GoRecv(doRecv func([]byte) error) {
	defer libol.Catch()
	libol.Info("TapWorker.GoRev")
	a.doRecv = doRecv

	for {
		data := make([]byte, a.ifmtu)
		if a.Ifce == nil {
			break
		}

		n, err := a.Ifce.Read(data)
		if err != nil {
			libol.Error("TapWorker.GoRev: %s", err)
			break
		}

		libol.Debug("TapWorker.GoRev: % x", data[:n])
		if a.Ifce.IsTUN() {
			eth := a.NewEth(libol.ETH_P_IP4)

			buffer := make([]byte, 0, a.ifmtu)
			buffer = append(buffer, eth.Encode()...)
			buffer = append(buffer, data[0:n]...)
			n += eth.Len

			doRecv(buffer[:n])
		} else {
			doRecv(data[:n])
		}
	}
	a.Close()
	libol.Warn("TapWorker.GoRev exit.")
}

func (a *TapWorker) DoSend(data []byte) error {
	libol.Debug("TapWorker.DoSend: % x", data)

	a.writechan <- data

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

		eth := a.NewEth(libol.ETH_P_ARP)

		reply := libol.NewArp()
		reply.OpCode = libol.ARP_REPLY
		reply.SIpAddr = a.EthSrcIp
		reply.TIpAddr = arp.SIpAddr
		reply.SHwAddr = a.EthSrcAddr
		reply.THwAddr = arp.SHwAddr

		buffer := make([]byte, 0, a.ifmtu)
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

func (a *TapWorker) GoLoop() {
	defer libol.Catch()
	libol.Info("TapWorker.GoLoop")

	for {
		select {
		case wdata := <-a.writechan:
			if a.Ifce == nil {
				break
			}

			if a.Ifce.IsTUN() {
				//Proxy arp request.
				if a.onArp(wdata) {
					libol.Info("TapWorker.GoLoop: Arp proxy.")
					continue
				}

				eth, err := libol.NewEtherFromFrame(wdata)
				if err != nil {
					libol.Error("TapWorker.GoLoop: %s", err)
					continue
				}
				if eth.IsVlan() {
					wdata = wdata[18:]
				} else if eth.IsIP4() {
					wdata = wdata[14:]
				} else { // default is Ethernet is 14 bytes.
					wdata = wdata[14:]
				}
			}

			if _, err := a.Ifce.Write(wdata); err != nil {
				libol.Error("TapWorker.GoLoop: %s", err)
			}
		}
	}
	a.Close()
	libol.Warn("TapWorker.GoLoop exit.")
}

func (a *TapWorker) Close() {
	libol.Info("TapWorker.Close")

	if a.Ifce != nil {
		a.Ifce.Close()
		a.Ifce = nil
	}
}

func (a *TapWorker) Stop() {
	a.Close()
}
