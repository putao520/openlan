package app

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/switch/storage"
)

type WithRequest struct {
	master Master
}

func NewWithRequest(m Master, c config.Switch) (r *WithRequest) {
	r = &WithRequest{
		master: m,
	}
	return
}

func (r *WithRequest) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	libol.Log("WithRequest.OnFrame %s.", frame)
	if !frame.IsControl() {
		return nil
	}
	action, body := frame.CmdAndParams()
	libol.Cmd("WithRequest.OnFrame: %s %s", action, body)
	switch action {
	case "neig=":
		r.OnNeighbor(client, body)
	case "ipad=":
		r.OnIpAddr(client, body)
	case "logi=":
		break
	default:
		r.OnDefault(client, body)
	}
	return nil
}

func (r *WithRequest) OnDefault(client libol.SocketClient, data string) {
	_ = client.WriteResp("pong", data)
}

func (r *WithRequest) OnNeighbor(client libol.SocketClient, data string) {
	libol.Info("TODO WithRequest.OnNeighbor: %s from %s", data, client)
}

func (r *WithRequest) OnIpAddr(client libol.SocketClient, data string) {
	libol.Info("WithRequest.OnIpAddr: %s from %s", data, client)

	n := models.NewNetwork("", "")
	if err := json.Unmarshal([]byte(data), n); err != nil {
		libol.Error("WithRequest.OnIpAddr: Invalid json data.")
		return
	}
	if n.Name == "" {
		n.Name = n.Tenant
	}
	if n.IfAddr == "" {
		FinNet := storage.Network.Get(n.Name)
		libol.Cmd("WithRequest.OnIpAddr: find %s", FinNet)
		uuid := storage.Point.GetUUID(client.Addr())
		ipStr, netmask := storage.Network.GetFreeAddr(uuid, FinNet)
		if ipStr == "" {
			libol.Error("WithRequest.OnIpAddr: %s no free address", n.Name)
			_ = client.WriteResp("ipaddr", "no free address")
			return
		}
		respNet := &models.Network{
			Name:    FinNet.Name,
			IfAddr:  ipStr,
			IpStart: ipStr,
			IpEnd:   ipStr,
			Netmask: netmask,
			Routes:  FinNet.Routes,
		}
		libol.Cmd("WithRequest.OnIpAddr: resp %s", respNet)
		if respStr, err := json.Marshal(respNet); err == nil {
			_ = client.WriteResp("ipaddr", string(respStr))
		}
		libol.Info("WithRequest.OnIpAddr: %s for %s", ipStr, client)
	}
}
