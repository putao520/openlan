package app

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/vswitch/service"
)

type WithRequest struct {
	master Master
}

func NewWithRequest(m Master, c config.VSwitch) (r *WithRequest) {
	r = &WithRequest{
		master: m,
	}
	return
}

func (r *WithRequest) OnFrame(client *libol.TcpClient, frame *libol.FrameMessage) error {
	libol.Debug("WithRequest.OnFrame %s.", frame)
	if frame.IsControl() {
		action, body := frame.CmdAndParams()
		libol.Debug("WithRequest.OnFrame: %s %s", action, body)

		switch action {
		case "neig=":
			r.OnNeighbor(client, body)
		case "ipad=":
			r.OnIpAddr(client, body)
		default:
			libol.Error("WithRequest.OnFrame: %s %s", action, body)
		}
	}

	return nil
}

func (r *WithRequest) OnNeighbor(client *libol.TcpClient, data string) {
	libol.Info("TODO WithRequest.OnNeighbor: %s from %s", data, client)
}

func (r *WithRequest) OnIpAddr(client *libol.TcpClient, data string) {
	libol.Info("WithRequest.OnIpAddr: %s from %s", data, client)

	net := models.NewNetwork("", "")
	if err := json.Unmarshal([]byte(data), net); err != nil {
		libol.Error("WithRequest.OnIpAddr: Invalid json data.")
		return
	}

	if net.IfAddr == "" {
		FinNet := service.Network.Get(net.Tenant)
		libol.Info("WithRequest.OnIpAddr: find %s", FinNet)
		ipStr, netmask := service.Network.GetFreeAddr(client, FinNet)
		if ipStr == "" {
			libol.Error("WithRequest.OnIpAddr: no free address")
			_ = client.WriteResp("ipaddr", "no free address")
			return
		}
		respNet := &models.Network{
			Tenant:  FinNet.Tenant,
			IfAddr:  ipStr,
			IpAddr:  ipStr,
			IpRange: 1,
			Netmask: netmask,
			Routes:  FinNet.Routes,
		}
		libol.Info("WithRequest.OnIpAddr: resp %s", respNet)
		if respStr, err := json.Marshal(respNet); err == nil {
			_ = client.WriteResp("ipaddr", string(respStr))
		}
	}
}
