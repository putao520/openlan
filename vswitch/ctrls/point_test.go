package ctrls

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/vswitch/service"
	"testing"
	"time"
)

func TestCtl_Point(t *testing.T) {
	libol.SetLog(libol.STACK)
	cc := &CtrlC{}
	err := cc.Open("http://localhost:10088/olan/upcall", "admin", "123")
	if err != nil {
		t.Error(err)
		return
	}
	cc.Start()
	point := models.Point{
		UUID:  "xxxx",
		Alias: "alias-test",
		Client: &libol.TcpClient{
			Addr: "xxx",
		},
	}
	service.Point.Add(&point)
	time.Sleep(5 * time.Second)
	service.Point.Del(point.Client.Addr)
	time.Sleep(5 * time.Second)
	cc.Stop()
}
