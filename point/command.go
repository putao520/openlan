package point

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
)

type Command struct {
	tcpWorker *TcpWorker
	brIp      net.IP
	brNet     *net.IPNet
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewCommand(config *config.Point) (cmd *Command) {
	var tlsConf *tls.Config
	if config.Tls {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}
	client := libol.NewTcpClient(config.Addr, tlsConf)
	cmd = &Command{}
	cmd.tcpWorker = NewTcpWorker(client, config, cmd)
	cmd.ctx, cmd.cancel = context.WithCancel(context.Background())
	return
}

func (cmd *Command) Connect() string {
	libol.Info("Command.Connect\n")

	if err := cmd.tcpWorker.Connect(); err != nil {
		return fmt.Sprintf("Command.Start %s", err)
	}

	return ""
}

func (cmd *Command) Start() {
	libol.Info("Command.Start\n")

	go cmd.tcpWorker.GoRecv(cmd.ctx, cmd.DoRecv)
	go cmd.tcpWorker.GoLoop(cmd.ctx)
}

func (cmd *Command) Close() {
	cmd.cancel()
	cmd.tcpWorker.Close()
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
	return cmd.tcpWorker.DoSend(data)
}

func (cmd *Command) DoOpen(args []string) string {
	//libol.Debug("Command.DoOpen %s", args)
	if len(args) > 0 {
		addr := args[0]
		config.RightAddr(&addr, 10002)

		cmd.tcpWorker.SetAddr(addr)
	}
	if len(args) > 1 {
		cmd.tcpWorker.SetAuth(args[1])
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
	arp.SHwAddr = libol.DEFAULTED
	arp.THwAddr = libol.ZEROED
	arp.SIpAddr = []byte(net.ParseIP(args[0]).To4())
	arp.TIpAddr = []byte(net.ParseIP(args[1]).To4())

	eth := libol.NewEther(libol.ETHPARP)
	eth.Dst = libol.BROADED
	eth.Src = libol.DEFAULTED

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
	ioReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("[point]# ")
		os.Stdout.Sync()

		cmdStr, err := ioReader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}

		input := strings.TrimSpace(strings.Trim(cmdStr, "\r\n"))
		if input == "quit" || input == "exit" {
			break
		}

		out := cmd.HitInput(strings.Split(input, " "))
		if out != "" {
			fmt.Println(out)
		}
	}
}

func (cmd *Command) OnIpAddr(worker *TcpWorker, n *models.Network) error {
	return nil
}