package main

import (
	"bufio"
	"fmt"
	"github.com/lightstar-dev/openlan-go/libol"
	"os"
	"strings"
)

type Server struct {

}

func (srv *Server) DoPoint(args []string) string {
	if len(args) <= 0 {
		return "verbose <level>"
	}

	fmt.Sscanf(args[0], "%d", libol.Log.Level)

	return fmt.Sprintf("%d", libol.Log.Level)
}

func (srv *Server) DovSwitch(args []string) string {
	if len(args) <= 0 {
		return "verbose <level>"
	}

	fmt.Sscanf(args[0], "%d", libol.Log.Level)

	return fmt.Sprintf("%d", libol.Log.Level)
}

func (srv *Server) HitInput(args []string) string {
	libol.Debug("Command.HitInput %s", args)

	switch args[0] {
	case "point":
		return srv.DoPoint(args[1:])
	case "vswitch":
		return srv.DovSwitch(args[1:])
	case "?":
		return "<command> [argument]..."
	}

	return ""
}

func (srv *Server) Loop() {
	ioReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("[openlan]# ")
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

		out := srv.HitInput(strings.Split(input, " "))
		if out != "" {
			fmt.Println(out)
		}
	}
}
