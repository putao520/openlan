package vswitch

import (
	"github.com/lightstar-dev/openlan-go/libol"
)

type WithRequest struct {
	worker WorkerApi
}

func NewWithRequest(api WorkerApi, c *Config) (w *WithRequest) {
	w = &WithRequest{
		worker: api,
	}
	return
}

func (w *WithRequest) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	libol.Debug("WithRequest.OnFrame % x.", frame.Data)

	if libol.IsInst(frame.Data) {
		action, body := libol.DecActionBody(frame.Data)
		libol.Debug("WithRequest.OnFrame.action: %s %s", action, body)

		if action == "neig=" {
			//TODO
		}
	}

	return nil
}
