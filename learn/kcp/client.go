package main

import (
	"github.com/xtaci/kcp-go/v5"
)

func main() {
	conn, err := kcp.DialWithOptions("192.168.209.141:10000", nil, 10, 3)
	if err!=nil {
		panic(err)
	}
	conn.Write([]byte("hello kcp.emmmmmmmmmmmmmmm"))
	select {}
}
