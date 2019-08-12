package main 


import (
	"fmt"
)

func main() {
    var input string

    for {
        fmt.Println("Please press enter `q` to exit...")
        fmt.Scanln(&input)
        if input == "q" {
            break
        }
	}
		
	fmt.Println("Done!")
}