package point

import (
	"fmt"
	"net"

	"github.com/lightstar-dev/openlan-go/libol"
)

type PointCmd struct {
	tcpwroker *TcpWorker
	brip      net.IP
	brnet     *net.IPNet
}

func NewPointCmd(config *Config) (this *PointCmd) {
	client := libol.NewTcpClient(config.Addr)

	this = &PointCmd{
		tcpwroker: NewTcpWorker(client, config),
	}
	return
}

func (this *PointCmd) Connect() string {
	libol.Info("PointCmd.Connect\n")

	if err := this.tcpwroker.Connect(); err != nil {
		return fmt.Sprintf("PointCmd.Start %s", err)
	}

	return ""
}

func (this *PointCmd) Start() {
	libol.Info("PointCmd.Start\n")

	go this.tcpwroker.GoRecv(this.DoRecv)
	go this.tcpwroker.GoLoop()
}

func (this *PointCmd) Close() {
	this.tcpwroker.Close()
}

func (this *PointCmd) onArp(data []byte) {
	libol.Debug("PointCmd.onArp\n")
	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		libol.Warn("PointCmd.onArp %s\n", err)
		return
	}

	if !eth.IsArp() {
		return
	}

	arp, err := libol.NewArpFromFrame(data[eth.Len:])
	if err != nil {
		libol.Error("PointCmd.onArp %s.", err)
		return
	}

	if arp.IsIP4() {
		if arp.OpCode != libol.ARP_REQUEST && arp.OpCode != libol.ARP_REPLY {
			return
		}

		libol.Info("PointCmd.onArp: %s on %s", net.HardwareAddr(arp.SHwAddr), net.IP(arp.SIpAddr))
	}
}

func (this *PointCmd) onIp4(data []byte) {
	//TODO
}

func (this *PointCmd) onStp(data []byte) {
	//TODO
}

func (this *PointCmd) DoRecv(data []byte) error {
	libol.Debug("PointCmd.DoRecv: % x\n", data)

	this.onArp(data)
	this.onStp(data)
	this.onIp4(data)

	return nil
}

func (this *PointCmd) DoSend(data []byte) error {
	return this.tcpwroker.DoSend(data)
}

func (this *PointCmd) DoOpen(args []string) string {
	//libol.Debug("PointCmd.DoOpen %s\n", args)
	if len(args) > 0 {
		addr := args[0]
		RightAddr(&addr, 10002)

		this.tcpwroker.SetAddr(addr)
	}
	if len(args) > 1 {
		this.tcpwroker.SetAuth(args[1])
	}

	return this.Connect()
}

// arp <source> <destination>
func (this *PointCmd) DoArp(args []string) string {
	libol.Debug("PointCmd.DoArp %s\n", args)
	if len(args) != 2 {
		return "arp <source> <destination>"
	}

	arp := libol.NewArp()
	arp.SHwAddr = libol.DEFAULTETHADDR
	arp.THwAddr = libol.ZEROETHADDR
	arp.SIpAddr = []byte(net.ParseIP(args[0]).To4())
	arp.TIpAddr = []byte(net.ParseIP(args[1]).To4())

	eth := libol.NewEther(libol.ETH_P_ARP)
	eth.Dst = libol.BROADETHADDR
	eth.Src = libol.DEFAULTETHADDR

	buffer := make([]byte, 0, 1024)
	buffer = append(buffer, eth.Encode()...)
	buffer = append(buffer, arp.Encode()...)
	this.DoSend(buffer)

	return fmt.Sprintf("%d", len(buffer))
}

func (this *PointCmd) DoClose(args []string) string {
	this.Close()
	return ""
}

func (this *PointCmd) DoVerbose(args []string) string {
	if len(args) <= 0 {
		return "verbose <level>"
	}

	fmt.Sscanf(args[0], "%d", libol.Log.Level)

	return fmt.Sprintf("%d", libol.Log.Level)
}

func (this *PointCmd) HitInput(args []string) string {
	//libol.Debug("PointCmd.HitInput %s\n", args)
	switch args[0] {
	case "open":
		return this.DoOpen(args[1:])
	case "close":
		return this.DoClose(args[1:])
	case "verbose":
		return this.DoVerbose(args[1:])
	case "arp":
		return this.DoArp(args[1:])
	case "?":
		return "<command> [argument]..."
	}

	return ""
}
