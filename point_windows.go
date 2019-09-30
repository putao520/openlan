package main

import (
    "fmt"
    "github.com/lightstar-dev/openlan-go/libol"
    "github.com/lightstar-dev/openlan-go/point"
)

func main() {
    c := point.NewConfig()
    libol.Debug("main.config: %s", c)
    p := point.NewPoint(c)

    p.Start()

    for {
        fmt.Println("Please press enter `q` to exit...")

        input := ""
        if fmt.Scanln(&input); input == "q" {
            break
        }
    }

    p.Close()
    fmt.Println("Done!")
}
