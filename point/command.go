package point

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/lightstar-dev/openlan-go/libol"
)

type Command struct {
	tcpwroker *TcpWorker
	brip      net.IP
	brnet     *net.IPNet
}

func NewCommand(config *Config) (cmd *Command) {
	client := libol.NewTcpClient(config.Addr)

	cmd = &Command{
		tcpwroker: NewTcpWorker(client, config),
	}
	return
}

func (cmd *Command) Connect() string {
	libol.Info("Command.Connect\n")

	if err := cmd.tcpwroker.Connect(); err != nil {
		return fmt.Sprintf("Command.Start %s", err)
	}

	return ""
}

func (cmd *Command) Start() {
	libol.Info("Command.Start\n")

	go cmd.tcpwroker.GoRecv(cmd.DoRecv)
	go cmd.tcpwroker.GoLoop()
}

func (cmd *Command) Close() {
	cmd.tcpwroker.Close()
}

func (cmd *Command) onArp(data []byte) {
	libol.Debug("Command.onArp\n")
	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		libol.Warn("Command.onArp %s", err)
		return
	}

	if !eth.IsArp() {
		return
	}

	arp, err := libol.NewArpFromFrame(data[eth.Len:])
	if err != nil {
		libol.Error("Command.onArp %s.", err)
		return
	}

	if arp.IsIP4() {
		if arp.OpCode != libol.ARP_REQUEST && arp.OpCode != libol.ARP_REPLY {
			return
		}

		libol.Info("Command.onArp: %s on %s", net.HardwareAddr(arp.SHwAddr), net.IP(arp.SIpAddr))
	}
}

func (cmd *Command) onIp4(data []byte) {
	//TODO
}

func (cmd *Command) onStp(data []byte) {
	//TODO
}

func (cmd *Command) DoRecv(data []byte) error {
	libol.Debug("Command.DoRecv: % x", data)

	cmd.onArp(data)
	cmd.onStp(data)
	cmd.onIp4(data)

	return nil
}

func (cmd *Command) DoSend(data []byte) error {
	return cmd.tcpwroker.DoSend(data)
}

func (cmd *Command) DoOpen(args []string) string {
	//libol.Debug("Command.DoOpen %s", args)
	if len(args) > 0 {
		addr := args[0]
		RightAddr(&addr, 10002)

		cmd.tcpwroker.SetAddr(addr)
	}
	if len(args) > 1 {
		cmd.tcpwroker.SetAuth(args[1])
	}

	return cmd.Connect()
}

// arp <source> <destination>
func (cmd *Command) DoArp(args []string) string {
	libol.Debug("Command.DoArp %s", args)
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

	//Send 3 times and thinks arp be ignored.
	cmd.DoSend(buffer)
	cmd.DoSend(buffer)
	cmd.DoSend(buffer)

	return fmt.Sprintf("%d", len(buffer))
}

func (cmd *Command) DoClose(args []string) string {
	cmd.Close()
	return ""
}

func (cmd *Command) DoVerbose(args []string) string {
	if len(args) <= 0 {
		return "verbose <level>"
	}

	fmt.Sscanf(args[0], "%d", libol.Log.Level)

	return fmt.Sprintf("%d", libol.Log.Level)
}

func (cmd *Command) HitInput(args []string) string {
	//libol.Debug("Command.HitInput %s", args)
	switch args[0] {
	case "open":
		return cmd.DoOpen(args[1:])
	case "close":
		return cmd.DoClose(args[1:])
	case "verbose":
		return cmd.DoVerbose(args[1:])
	case "arp":
		return cmd.DoArp(args[1:])
	case "?":
		return "<command> [argument]..."
	}

	return ""
}

func (cmd *Command) Loop() {
	ioreader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("[point]# ")
		os.Stdout.Sync()

		cmdstr, err := ioreader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}

		input := strings.TrimSpace(strings.Trim(cmdstr, "\r\n"))
		if input == "quit" || input == "exit" {
			break
		}

		out := cmd.HitInput(strings.Split(input, " "))
		if out != "" {
			fmt.Println(out)
		}
	}
}
