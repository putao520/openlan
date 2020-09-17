package olap

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/network"
	"io"
	"os"
	"strings"
)

type Pointer interface {
	Status() string
	Addr() string
	IfName() string
	IfAddr() string
	Client() libol.SocketClient
	Device() network.Taper
	UpTime() int64
	UUID() string
	User() string
}

type MixPoint struct {
	// private
	uuid   string
	worker *Worker
	config *config.Point
	out    *libol.SubLogger
}

func NewMixPoint(config *config.Point) MixPoint {
	p := MixPoint{
		worker: NewWorker(config),
		config: config,
		out:    libol.NewSubLogger(config.Id()),
	}
	return p
}

func (p *MixPoint) Initialize() {
	libol.Info("MixPoint.Initialize")
	p.worker.SetUUID(p.UUID())
	p.worker.Initialize()
}

func (p *MixPoint) UUID() string {
	if p.uuid == "" {
		p.uuid = libol.GenToken(32)
	}
	return p.uuid
}

func (p *MixPoint) Status() libol.SocketStatus {
	return p.worker.Status()
}

func (p *MixPoint) Addr() string {
	return p.worker.Addr()
}

func (p *MixPoint) IfName() string {
	return p.worker.IfName()
}

func (p *MixPoint) Client() libol.SocketClient {
	return p.worker.Client()
}

func (p *MixPoint) Device() network.Taper {
	return p.worker.Device()
}

func (p *MixPoint) UpTime() int64 {
	return p.worker.UpTime()
}

func (p *MixPoint) IfAddr() string {
	return p.worker.ifAddr
}

func (p *MixPoint) Tenant() string {
	if p.config != nil {
		return p.config.Network
	}
	return ""
}

func (p *MixPoint) User() string {
	if p.config != nil {
		return p.config.Username
	}
	return ""
}

func (p *MixPoint) Record() map[string]int64 {
	rt := p.worker.conWorker.record
	// TODO padding data from tapWorker
	return rt.Data()
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItem("show",
		readline.PcItem("record"),
		readline.PcItem("statistics"),
	),
	readline.PcItem("bye"),
	readline.PcItem("help"),
)

var cfg = &readline.Config{
	Prompt:            ">>",
	HistoryFile:       ".history",
	AutoComplete:      completer,
	InterruptPrompt:   "^C",
	EOFPrompt:         "quit",
	HistorySearchFold: true,
	VimMode:           true,
	FuncFilterInputRune: func(r rune) (rune, bool) {
		switch r {
		// block CtrlZ feature
		case readline.CharCtrlZ:
			return r, false
		}
		return r, true
	},
}

func (p *MixPoint) CmdShow(args string) {
	switch args {
	case "record":
		fmt.Printf("%-15s: %v\n", "Record", p.Record())
	case "statistics":
		if c := p.Client(); c != nil {
			fmt.Printf("%-15s: %v\n", "Statistics", c.Statistics())
		}
	default:
		c := p.Client()
		fmt.Printf("%-15s: %s\n", "UUID", p.UUID())
		fmt.Printf("%-15s: %d\n", "UpTime", p.UpTime())
		fmt.Printf("%-15s: %s\n", "Device", p.IfName())
		if c != nil {
			fmt.Printf("%-15s: %s\n", "Status", c.Status())
		}
	}
}

func (p *MixPoint) ReadLine() {
	cfg.Prompt = "[" + p.config.Alias + "@olap]# "
	l, err := readline.NewEx(cfg)
	if err != nil {
		p.out.Error("Point.ReadLine %s", err)
	}
	defer l.Close()
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "mode "):
			switch line[5:] {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			}
		case line == "show", line == "":
			p.CmdShow("")
		case line == "bye", line == "quit":
			if proc, err := os.FindProcess(os.Getpid()); err == nil {
				_ = proc.Signal(os.Kill)
			} else {
				p.out.Error("Point.ReadLine %s", err)
			}
			goto quit
		case strings.HasPrefix(line, "show "):
			p.CmdShow(line[5:])
		}
	}
quit:
}
