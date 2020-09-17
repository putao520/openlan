package olap

import (
	"encoding/json"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/danieldin95/openlan-go/src/libol"
	"io"
	"os"
	"strings"
)

type Terminal struct {
	Pointer   Pointer
	Completer readline.AutoCompleter
	Config    *readline.Config
}

func NewTerminal(pointer Pointer) *Terminal {
	t := &Terminal{Pointer: pointer}
	t.Completer = readline.NewPrefixCompleter(
		readline.PcItem("bye"),
		readline.PcItem("help"),
		readline.PcItem("mode",
			readline.PcItem("vi"),
			readline.PcItem("emacs"),
		),
		readline.PcItem("show",
			readline.PcItem("config"),
			readline.PcItem("record"),
			readline.PcItem("statistics"),
		),
		readline.PcItem("edit",
			readline.PcItem("user"),
			readline.PcItem("connection"),
		),
	)
	t.Config = &readline.Config{
		Prompt:            t.Prompt(),
		HistoryFile:       ".history",
		InterruptPrompt:   "^C",
		EOFPrompt:         "quit",
		HistorySearchFold: true,
		AutoComplete:      t.Completer,
	}
	return t
}

func (t *Terminal) Prompt() string {
	user := t.Pointer.User()
	alias := t.Pointer.Alias()
	tenant := t.Pointer.Tenant()
	return fmt.Sprintf("[%s@%s %s]# ", user, alias, tenant)
}

func (t *Terminal) CmdEdit(args string) {
}

func (t *Terminal) CmdShow(args string) {
	switch args {
	case "record":
		fmt.Printf("%-15s: %v\n", "Record", t.Pointer.Record())
	case "statistics":
		if c := t.Pointer.Client(); c != nil {
			fmt.Printf("%-15s: %v\n", "Statistics", c.Statistics())
		}
	case "config":
		cfg := t.Pointer.Config()
		if str, err := json.MarshalIndent(cfg, "", "  "); err == nil {
			fmt.Printf("%s\n", str)
		} else {
			fmt.Printf("Point.CmdShow %s\n", err)
		}
	default:
		c := t.Pointer.Client()
		fmt.Printf("%-15s: %s\n", "UUID", t.Pointer.UUID())
		fmt.Printf("%-15s: %d\n", "UpTime", t.Pointer.UpTime())
		fmt.Printf("%-15s: %s\n", "Device", t.Pointer.IfName())
		if c != nil {
			fmt.Printf("%-15s: %s\n", "Status", c.Status())
		}
	}
}

func (t *Terminal) Trim(v string) string {
	return strings.TrimSpace(v)
}

func (t *Terminal) Start() {
	l, err := readline.NewEx(t.Config)
	if err != nil {
		libol.Error("Terminal.ReadLine %s", err)
		return
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
		line = t.Trim(line)
		switch {
		case strings.HasPrefix(line, "mode "):
			switch t.Trim(line[5:]) {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			}
		case line == "show":
			t.CmdShow("")
		case line == "bye", line == "quit":
			if proc, err := os.FindProcess(os.Getpid()); err == nil {
				_ = proc.Signal(os.Interrupt)
			} else {
				fmt.Printf("Terminal.ReadLine %s", err)
			}
			goto quit
		case strings.HasPrefix(line, "show "):
			t.CmdShow(t.Trim(line[5:]))
		case strings.HasPrefix(line, "edit "):
			t.CmdEdit(t.Trim(line[5:]))
		}
	}
quit:
}
