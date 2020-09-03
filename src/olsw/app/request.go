package app

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/olsw/schema"
	"github.com/danieldin95/openlan-go/src/olsw/storage"
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
	out := client.Out()
	if frame.IsEthernet() {
		return nil
	}
	if out.Has(libol.DEBUG) {
		out.Log("WithRequest.OnFrame %s.", frame)
	}
	action, body := frame.CmdAndParams()
	if out.Has(libol.CMD) {
		out.Cmd("WithRequest.OnFrame: %s %s", action, body)
	}
	switch action {
	case libol.NeighborReq:
		r.OnNeighbor(client, body)
	case libol.IpAddrReq:
		r.OnIpAddr(client, body)
	case libol.LeftReq:
		r.OnLeave(client, body)
	case libol.LoginReq:
		out.Debug("WithRequest.OnFrame %s: %s", action, body)
	default:
		r.OnDefault(client, body)
	}
	return nil
}

func (r *WithRequest) OnDefault(client libol.SocketClient, data []byte) {
	m := libol.NewControlFrame(libol.PongResp, data)
	_ = client.WriteMsg(m)
}

func (r *WithRequest) OnNeighbor(client libol.SocketClient, data []byte) {
	resp := make([]schema.Neighbor, 0, 32)
	for obj := range storage.Neighbor.List() {
		if obj == nil {
			break
		}
		resp = append(resp, models.NewNeighborSchema(obj))
	}
	if respStr, err := json.Marshal(resp); err == nil {
		m := libol.NewControlFrame(libol.NeighborResp, respStr)
		_ = client.WriteMsg(m)
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
				lease.Alias = alias
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
		lease.Client = p.Client.Address()
	}
	return lease
}
func (r *WithRequest) OnIpAddr(client libol.SocketClient, data []byte) {
	var resp *models.Network
	out := client.Out()
	out.Info("WithRequest.OnIpAddr: %s", data)
	recv := models.NewNetwork("", "")
	if err := json.Unmarshal(data, recv); err != nil {
		out.Error("WithRequest.OnIpAddr: invalid json data.")
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
		out.Error("WithRequest.OnIpAddr: invalid network %s.", recv.Name)
		return
	}
	out.Cmd("WithRequest.OnIpAddr: find %s", n)
	p := storage.Point.Get(client.Address())
	if p == nil {
		out.Error("WithRequest.OnIpAddr: point notFound")
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
		out.Cmd("WithRequest.OnIpAddr: resp %s", resp)
		if respStr, err := json.Marshal(resp); err == nil {
			m := libol.NewControlFrame(libol.IpAddrResp, respStr)
			_ = client.WriteMsg(m)
		}
		out.Info("WithRequest.OnIpAddr: %s", resp.IfAddr)
	} else {
		out.Error("WithRequest.OnIpAddr: %s no free address", recv.Name)
		m := libol.NewControlFrame(libol.IpAddrResp, []byte("no free address"))
		_ = client.WriteMsg(m)
	}
}

func (r *WithRequest) OnLeave(client libol.SocketClient, data []byte) {
	out := client.Out()
	out.Info("WithRequest.OnLeave")
	r.master.OffClient(client)
}
