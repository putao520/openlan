package point

import (
	"net"

	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/songgao/water"
)

type TapWorker struct {
	ifce      *water.Interface
	writechan chan []byte
	ifmtu     int
	doRecv    func([]byte) error

	//for tunnel device.
	EthDstAddr []byte
	EthSrcAddr []byte
	EthSrcIp   []byte
}

func NewTapWorker(ifce *water.Interface, c *Config) (this *TapWorker) {
	this = &TapWorker{
		ifce:      ifce,
		writechan: make(chan []byte, 1024*10),
		ifmtu:     c.Ifmtu, //1514
	}

	if this.ifce.IsTUN() {
		this.EthSrcIp = net.ParseIP(c.Ifaddr).To4()
		libol.Info("NewTapWoker srcIp: % x\n", this.EthSrcIp)

		if c.Ifethsrc == "" {
			this.EthSrcAddr = libol.GenEthAddr(6)
		} else {
			if hw, err := net.ParseMAC(c.Ifethsrc); err == nil {
				this.EthSrcAddr = []byte(hw)
			}
		}
		if hw, err := net.ParseMAC(c.Ifethdst); err == nil {
			this.EthDstAddr = []byte(hw)
		}
		libol.Info("NewTapWorker src: % x, dst: % x\n", this.EthSrcAddr, this.EthDstAddr)
	}

	return
}

func (this *TapWorker) NewEth(t uint16) *libol.Ether {
	eth := libol.NewEther(t)
	eth.Dst = this.EthDstAddr
	eth.Src = this.EthSrcAddr

	return eth
}

func (this *TapWorker) GoRecv(doRecv func([]byte) error) {
	this.doRecv = doRecv
	defer this.Close()
	for {
		data := make([]byte, this.ifmtu)
		if this.ifce == nil {
			break
		}

		n, err := this.ifce.Read(data)
		if err != nil {
			libol.Error("TapWorker.GoRev: %s", err)
			break
		}

		libol.Debug("TapWorker.GoRev: % x\n", data[:n])
		if this.ifce.IsTUN() {
			eth := this.NewEth(libol.ETH_P_IP4)

			buffer := make([]byte, 0, this.ifmtu)
			buffer = append(buffer, eth.Encode()...)
			buffer = append(buffer, data[0:n]...)
			n += eth.Len

			doRecv(buffer[:n])
		} else {
			doRecv(data[:n])
		}
	}
}

func (this *TapWorker) DoSend(data []byte) error {
	libol.Debug("TapWorker.DoSend: % x\n", data)

	this.writechan <- data

	return nil
}

func (this *TapWorker) onArp(data []byte) bool {
	libol.Debug("TapWorker.onArp\n")
	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		libol.Warn("TapWorker.onArp %s\n", err)
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

		eth := this.NewEth(libol.ETH_P_ARP)

		reply := libol.NewArp()
		reply.OpCode = libol.ARP_REPLY
		reply.SIpAddr = this.EthSrcIp
		reply.TIpAddr = arp.SIpAddr
		reply.SHwAddr = this.EthSrcAddr
		reply.THwAddr = arp.SHwAddr

		buffer := make([]byte, 0, this.ifmtu)
		buffer = append(buffer, eth.Encode()...)
		buffer = append(buffer, reply.Encode()...)

		libol.Info("TapWorker.onArp % x.", buffer)
		if this.doRecv != nil {
			this.doRecv(buffer)
		}

		return true
	}

	return false
}

func (this *TapWorker) GoLoop() {
	defer this.Close()
	for {
		select {
		case wdata := <-this.writechan:
			if this.ifce == nil {
				return
			}

			if this.ifce.IsTUN() {
				//Proxy arp request.
				if this.onArp(wdata) {
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

			if _, err := this.ifce.Write(wdata); err != nil {
				libol.Error("TapWorker.GoLoop: %s", err)
			}
		}
	}
}

func (this *TapWorker) Close() {
	libol.Info("TapWorker.Close")

	if this.ifce != nil {
		this.ifce.Close()
		this.ifce = nil
	}
}
