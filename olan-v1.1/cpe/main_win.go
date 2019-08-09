package main

import (
	"log"
	"flag"
	"fmt"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"

	"./openlan"
)

func ifceLoop(client *openlan.TcpClient, ifce *water.Interface) {
	var frame ethernet.Frame

	for {
		frame.Resize(1500)
		n, err := ifce.Read([]byte(frame))
		if err != nil {
			log.Fatal(err)
		}

		frame = frame[:n]
		log.Printf("i--Dst: %s\n", frame.Destination())
		log.Printf("i--Src: %s\n", frame.Source())
		log.Printf("i--Ethertype: % x\n", frame.Ethertype())
		log.Printf("i--Payload: % x\n", frame.Payload())

		if err := client.SendMsg([]byte(frame)); err != nil {
			log.Fatal(err)
		}
	}
}

func clientLoop(client *openlan.TcpClient, ifce *water.Interface) {
	var frame ethernet.Frame

	for {
		frame.Resize(1500)
		n, err := client.RecvMsg([]byte(frame))
		if err != nil {
			log.Fatal(err)
		}

		frame = frame[:n]
		log.Printf("o--Dst: %s\n", frame.Destination())
		log.Printf("o--Src: %s\n", frame.Source())
		log.Printf("o--Ethertype: % x\n", frame.Ethertype())
		log.Printf("o--Payload: % x\n", frame.Payload())
		
		n, err = ifce.Write([]byte(frame))
		if err != nil {
			log.Fatal(err)
		}		
	}
}

func main() {
	var addr string
	var port *int

	flag.StringVar(&addr, "addr", "openlan.net",  "the server address")
	port = flag.Int("port", 10001, "the port number")

	flag.Parse()

	ifce, err := water.New(water.Config{
		DeviceType: water.TAP,
	})
	if err != nil {
		log.Fatal(err)
	}

	client, err:= openlan.NewTcpClient(addr, uint16(*port))
	if err != nil {
		log.Printf("connect failed: %s.", err);
	}

	go ifceLoop(client, ifce)
	go clientLoop(client, ifce)

	fmt.Scanln()
	fmt.Println("press enter to exit...")
	fmt.Scanln()
    fmt.Println("done")
}
