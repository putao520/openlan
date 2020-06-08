package ctrls

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/vswitch/storage"
	"testing"
	"time"
)

func TestCtl_Point(t *testing.T) {
	libol.SetLog(libol.STACK)
	cc := &CtrlC{
		Url:      "http://localhost:10088/ctrl",
		Name:     "admin",
		Password: "123",
	}
	err := cc.Open()
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
	storage.Point.Add(&point)
	time.Sleep(5 * time.Second)
	storage.Point.Del(point.Client.Addr)
	time.Sleep(5 * time.Second)
	cc.Stop()
	time.Sleep(5 * time.Second)
}
