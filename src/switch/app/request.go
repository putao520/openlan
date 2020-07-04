package app

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/switch/schema"
	"github.com/danieldin95/openlan-go/src/switch/storage"
	"strings"
)

type WithRequest struct {
	master Master
}

func NewWithRequest(m Master, c config.Switch) *WithRequest {
	return &WithRequest{
		master: m,
	}
}

func (r *WithRequest) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	if frame.IsEthernet() {
		return nil
	}
	if libol.HasLog(libol.DEBUG) {
		libol.Log("WithRequest.OnFrame %s.", frame)
	}
	action, body := frame.CmdAndParams()
	if libol.HasLog(libol.CMD) {
		libol.Cmd("WithRequest.OnFrame: %s %s", action, body)
	}
	switch action {
	case "neig=":
		r.OnNeighbor(client, body)
	case "ipad=":
		r.OnIpAddr(client, body)
	case "left=":
		r.OnLeave(client, body)
	case "logi=":
		libol.Debug("WithRequest.OnFrame %s: %s", action, body)
	default:
		r.OnDefault(client, body)
	}
	return nil
}

func (r *WithRequest) OnDefault(client libol.SocketClient, data string) {
	_ = client.WriteResp("pong", data)
}

func (r *WithRequest) OnNeighbor(client libol.SocketClient, data string) {
	resp := make([]schema.Neighbor, 0, 32)
	for obj := range storage.Neighbor.List() {
		if obj == nil {
			break
		}
		resp = append(resp, models.NewNeighborSchema(obj))
	}
	if respStr, err := json.Marshal(resp); err == nil {
		_ = client.WriteResp("neighbor", string(respStr))
	}
}

func (r *WithRequest) OnIpAddr(client libol.SocketClient, data string) {
	libol.Info("WithRequest.OnIpAddr: %s from %s", data, client)

	rcvNet := models.NewNetwork("", "")
	if err := json.Unmarshal([]byte(data), rcvNet); err != nil {
		libol.Error("WithRequest.OnIpAddr: Invalid json data.")
		return
	}
	if rcvNet.Name == "" {
		rcvNet.Name = rcvNet.Tenant
	}
	net := storage.Network.Get(rcvNet.Name)
	libol.Cmd("WithRequest.OnIpAddr: find %s", net)
	uuid := storage.Point.GetUUID(client.Addr())

	var resp *models.Network
	if rcvNet.IfAddr == "" {
		ipStr, netmask := storage.Network.GetFreeAddr(uuid, net)
		if ipStr != "" {
			resp = &models.Network{
				Name:    net.Name,
				IfAddr:  ipStr,
				IpStart: ipStr,
				IpEnd:   ipStr,
				Netmask: netmask,
				Routes:  net.Routes,
			}
		}
	} else {
		ipAddr := strings.SplitN(rcvNet.IfAddr, "/", 2)[0]
		storage.Network.AddUsedAddr(uuid, ipAddr)
		resp = rcvNet
	}
	if resp != nil {
		libol.Cmd("WithRequest.OnIpAddr: resp %s", resp)
		if respStr, err := json.Marshal(resp); err == nil {
			_ = client.WriteResp("ipaddr", string(respStr))
		}
		libol.Info("WithRequest.OnIpAddr: %s for %s", resp.IfAddr, client)
	} else {
		libol.Error("WithRequest.OnIpAddr: %s no free address", rcvNet.Name)
		_ = client.WriteResp("ipaddr", "no free address")
	}
}

func (r *WithRequest) OnLeave(client libol.SocketClient, data string) {
	libol.Info("WithRequest.OnLeave: %s", client.RemoteAddr())
	r.master.OffClient(client)
}
