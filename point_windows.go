package main

import (
    "fmt"

    "github.com/lightstar-dev/openlan-go/point"
    "github.com/lightstar-dev/openlan-go/point"
)

func main() {
    c := point.NewConfig()
    libol.Debug("main.config: %s", c)
    p := point.NewPoint(c)

    p.Start()

    for {
        var input string

        fmt.Println("Please press enter `q` to exit...")
        if fmt.Scanln(&input); input == "q" {
            break
        }
    }

    p.Close()
    fmt.Println("Done!")
}
