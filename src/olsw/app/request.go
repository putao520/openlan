package app

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/olsw/storage"
	"github.com/danieldin95/openlan-go/src/schema"
	"strings"
)

type Request struct {
	master Master
}

func NewRequest(m Master) *Request {
	return &Request{
		master: m,
	}
}

func (r *Request) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	out := client.Out()
	if frame.IsEthernet() {
		return nil
	}
	if out.Has(libol.DEBUG) {
		out.Log("Request.OnFrame %s.", frame)
	}
	action, body := frame.CmdAndParams()
	if out.Has(libol.CMD) {
		out.Cmd("Request.OnFrame: %s %s", action, body)
	}
	switch action {
	case libol.NeighborReq:
		r.onNeighbor(client, body)
	case libol.IpAddrReq:
		r.onIpAddr(client, body)
	case libol.LeftReq:
		r.onLeave(client, body)
	case libol.LoginReq:
		out.Debug("Request.OnFrame %s: %s", action, body)
	default:
		r.onDefault(client, body)
	}
	return nil
}

func (r *Request) onDefault(client libol.SocketClient, data []byte) {
	m := libol.NewControlFrame(libol.PongResp, data)
	_ = client.WriteMsg(m)
}

func (r *Request) onNeighbor(client libol.SocketClient, data []byte) {
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

func (r *Request) getLease(ifAddr string, p *models.Point, n *models.Network) *schema.Lease {
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
		lease.Client = p.Client.String()
	}
	return lease
}
func (r *Request) onIpAddr(client libol.SocketClient, data []byte) {
	var resp *models.Network
	out := client.Out()
	out.Info("Request.onIpAddr: %s", data)
	recv := models.NewNetwork("", "")
	if err := json.Unmarshal(data, recv); err != nil {
		out.Error("Request.onIpAddr: invalid json data.")
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
		out.Error("Request.onIpAddr: invalid network %s.", recv.Name)
		return
	}
	out.Cmd("Request.onIpAddr: find %s", n)
	p := storage.Point.Get(client.String())
	if p == nil {
		out.Error("Request.onIpAddr: point notFound")
		return
	}
	lease := r.getLease(recv.IfAddr, p, n)
	if recv.IfAddr == "" { // not interface address, and try to alloc it.
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
		// get release failed.
	} else {
		resp = recv
	}
	if resp != nil {
		out.Cmd("Request.onIpAddr: resp %s", resp)
		if respStr, err := json.Marshal(resp); err == nil {
			m := libol.NewControlFrame(libol.IpAddrResp, respStr)
			_ = client.WriteMsg(m)
		}
		out.Info("Request.onIpAddr: %s", resp.IfAddr)
	} else {
		out.Error("Request.onIpAddr: %s no free address", recv.Name)
		m := libol.NewControlFrame(libol.IpAddrResp, []byte("no free address"))
		_ = client.WriteMsg(m)
	}
}

func (r *Request) onLeave(client libol.SocketClient, data []byte) {
	out := client.Out()
	out.Info("Request.onLeave")
	r.master.OffClient(client)
}
