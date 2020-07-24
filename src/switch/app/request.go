package app

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
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

func (r *WithRequest) GetLease(ifAddr string, p *models.Point, n *models.Network) *schema.Lease {
	if n == nil {
		return nil
	}
	uuid := p.UUID
	alias := p.Alias
	network := n.Name
	lease := storage.Network.GetLeaseByAlias(alias) // try by alias firstly
	if ifAddr == "" {
		if lease == nil { // now to alloc it.
			lease = storage.Network.NewLease(uuid, network)
			if lease != nil {
				lease.Alias  = alias
			}
		} else {
			lease.UUID = uuid
		}
	} else {
		ipAddr := strings.SplitN(ifAddr, "/", 2)[0]
		if lease != nil && lease.Address == ipAddr {
			lease.UUID = uuid
		}
		if lease == nil || lease.Address != ipAddr {
			lease = storage.Network.AddLease(uuid, ipAddr)
			lease.Alias = alias
		}
	}
	if lease != nil {
		lease.Network = network
		lease.Client = p.Client.Addr()
	}
	return lease
}
func (r *WithRequest) OnIpAddr(client libol.SocketClient, data string) {
	var resp *models.Network
	libol.Info("WithRequest.OnIpAddr: %s from %s", data, client)
	recv := models.NewNetwork("", "")
	if err := json.Unmarshal([]byte(data), recv); err != nil {
		libol.Error("WithRequest.OnIpAddr: invalid json data.")
		return
	}
	if recv.Name == "" {
		recv.Name = recv.Tenant
	}
	if recv.Name == "" {
		recv.Name = "default"
	}
	n := storage.Network.Get(recv.Name)
	if n == nil {
		libol.Error("WithRequest.OnIpAddr: invalid network %s.", recv.Name)
		return
	}
	libol.Cmd("WithRequest.OnIpAddr: find %s", n)
	p := storage.Point.Get(client.Addr())
	if p == nil {
		libol.Error("WithRequest.OnIpAddr: invalid point %s.", client)
		return
	}
	lease := r.GetLease(recv.IfAddr, p, n)
	if recv.IfAddr == "" { // If not configure interface address, and try to alloc it.
		if lease != nil {
			resp = &models.Network{
				Name:    n.Name,
				IfAddr:  lease.Address,
				IpStart: n.IpStart,
				IpEnd:   n.IpEnd,
				Netmask: n.Netmask,
				Routes:  n.Routes,
			}
		}
	} else {
		resp = recv
	}
	if resp != nil {
		libol.Cmd("WithRequest.OnIpAddr: resp %s", resp)
		if respStr, err := json.Marshal(resp); err == nil {
			_ = client.WriteResp("ipaddr", string(respStr))
		}
		libol.Info("WithRequest.OnIpAddr: %s for %s", resp.IfAddr, client)
	} else {
		libol.Error("WithRequest.OnIpAddr: %s no free address", recv.Name)
		_ = client.WriteResp("ipaddr", "no free address")
	}
}

func (r *WithRequest) OnLeave(client libol.SocketClient, data string) {
	libol.Info("WithRequest.OnLeave: %s", client.RemoteAddr())
	r.master.OffClient(client)
}
