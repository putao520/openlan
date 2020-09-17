package olap

import (
	"encoding/json"
	"fmt"
	"github.com/chzyer/readline"
	"io"
	"strings"
)

type Terminal struct {
	Pointer Pointer
	Console *readline.Instance
}

func NewTerminal(pointer Pointer) *Terminal {
	t := &Terminal{Pointer: pointer}
	completer := readline.NewPrefixCompleter(
		readline.PcItem("quit"),
		readline.PcItem("help"),
		readline.PcItem("mode",
			readline.PcItem("vi"),
			readline.PcItem("emacs"),
		),
		readline.PcItem("show",
			readline.PcItem("config"),
			readline.PcItem("network"),
			readline.PcItem("record"),
			readline.PcItem("statistics"),
		),
		readline.PcItem("edit",
			readline.PcItem("user"),
			readline.PcItem("connection"),
		),
	)

	config := &readline.Config{
		Prompt:            t.Prompt(),
		HistoryFile:       ".history",
		InterruptPrompt:   "^C",
		EOFPrompt:         "quit",
		HistorySearchFold: true,
		AutoComplete:      completer,
	}
	if l, err := readline.NewEx(config); err == nil {
		t.Console = l
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
	case "network":
		cfg := t.Pointer.Network()
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

func (t *Terminal) CmdBye() {
}

func (t *Terminal) CmdMode(args string) {
	switch args {
	case "vi":
		t.Console.SetVimMode(true)
	case "emacs":
		t.Console.SetVimMode(false)
	}
}

func (t *Terminal) Start() {
	if t.Console == nil {
		return
	}
	defer t.Console.Close()
	for {
		line, err := t.Console.Readline()
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
			t.CmdMode(t.Trim(line[5:]))
		case line == "show":
			t.CmdShow("")
		case line == "quit":
			t.CmdBye()
			goto quit
		case strings.HasPrefix(line, "show "):
			t.CmdShow(t.Trim(line[5:]))
		case strings.HasPrefix(line, "edit "):
			t.CmdEdit(t.Trim(line[5:]))
		}
	}
quit:
	fmt.Printf("Terminal.Start quit")
}
