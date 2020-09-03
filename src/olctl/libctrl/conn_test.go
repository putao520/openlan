package libctrl

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"testing"
	"time"
)

func TestCtl_Conn_Open_Ctrl(t *testing.T) {
	ws := &libol.WsClient{
		Auth: libol.Auth{
			Type:     "basic",
			Username: "admin",
			Password: "123",
		},
		Url: "http://localhost:10088/ctrl",
	}
	ws.Initialize()
	to, err := ws.Dial()
	if err != nil {
		t.Error(err)
		return
	}
	defer to.Close()

	m := &Message{}
	if err := Codec.Receive(to, m); err != nil {
		t.Error(err)
	}
	fmt.Println(m)
	conn := CtrlConn{
		Conn: to,
	}
	conn.Open()
	conn.Start()
	conn.Send(Message{Resource: "hello", Data: "from client"})
	if err := conn.SendWait(Message{Resource: "hello", Data: "from client"}); err != nil {
		t.Error(err)
	}
	time.Sleep(5 * time.Second)
	conn.Stop()
	conn.Send(Message{Resource: "hello", Data: "from client"})
	time.Sleep(5 * time.Second)
}
