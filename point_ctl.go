package main

import (
    "os"
    "fmt"
    "bufio"
    "strings"

    "github.com/lightstar-dev/openlan-go/libol"
    "github.com/lightstar-dev/openlan-go/point"
)

func main() {
    c := point.NewConfig()
    libol.Debug("main.config: %s", c)

    p := point.NewPointCmd(c)
    p.Start()

    ioreader := bufio.NewReader(os.Stdin)

    for {
        fmt.Print("[point]# ")
        os.Stdout.Sync()

        cmdstr, err := ioreader.ReadString('\n'); 
        if err != nil {
            fmt.Println(err)
            break
        }

        input := strings.TrimSpace(strings.Trim(cmdstr, "\r\n"))
        if input == "quit" || input == "exit" {
            break
        }
 
        out := p.HitInput(strings.Split(input, " "))        
        if out != "" {
            fmt.Println(out)
        }
    }
}
