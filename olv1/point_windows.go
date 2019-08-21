package main

import (
    "fmt"
    "log"

    "github.com/danieldin95/openlan-go/olv1/cpe"
)

func main() {
    c := point.NewConfig()
    log.Printf("Debug| main.config: %s", c)
    cpe := point.NewCpe(c)

    cpe.Start()

    for {
        var input string

        fmt.Println("Please press enter `q` to exit...")
        if fmt.Scanln(&input); input == "q" {
            break
        }
    }

    cpe.Close()
    fmt.Println("Done!")
}
