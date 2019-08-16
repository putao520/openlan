package main

import (
    "fmt"
    "log"

    "github.com/danieldin95/openlan-go/olv1/cpe"
)

func main() {
    c := olv1cpe.NewConfig()
    log.Printf("Debug| main.config: %s", c)
    cpe := olv1cpe.NewCpe(c)

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
