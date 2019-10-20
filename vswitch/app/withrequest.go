package app

import (
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/vswitch/api"
)

type WithRequest struct {
	worker api.Worker
}

func NewWithRequest(w api.Worker, c *config.VSwitch) (r *WithRequest) {
	r = &WithRequest{
		worker: w,
	}
	return
}

func (r *WithRequest) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	libol.Debug("WithRequest.OnFrame % x.", frame.Data)

	if libol.IsInst(frame.Data) {
		action, body := libol.DecActionBody(frame.Data)
		libol.Debug("WithRequest.OnFrame.action: %s %s", action, body)

		switch action {
		case "neig=":
			//TODO
		default:
			//TODO
		}
	}

	return nil
}
