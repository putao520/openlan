package store

import (
	"bufio"
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/schema"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type _ovClient struct {
	Directory string
}

func ParseInt64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func (o *_ovClient) GetDevice(name string) string {
	sw := config.Manager.Switch
	if sw == nil {
		return ""
	}
	for _, n := range sw.Network {
		vpn := n.OpenVPN
		if vpn == nil {
			continue
		}
		if vpn.Network == name {
			return vpn.Device
		}
	}
	return ""
}

func (o *_ovClient) scanStatus(network string, reader io.Reader) (map[string]*schema.OvClient, error) {
	readAt := "header"
	offset := 0
	scanner := bufio.NewScanner(reader)
	clients := make(map[string]*schema.OvClient, 32)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "OpenVPN CLIENT LIST" {
			readAt = "common"
			offset = 3
		}
		if line == "ROUTING TABLE" {
			readAt = "routing"
			offset = 2
		}
		if line == "GLOBAL STATS" {
			readAt = "global"
			offset = 1
		}
		if offset > 0 {
			offset -= 1
			continue
		}
		columns := strings.SplitN(line, ",", 5)
		switch readAt {
		case "common":
			if len(columns) == 5 {
				name := columns[0]
				client := &schema.OvClient{
					Name:   columns[0],
					Remote: columns[1],
					State:  "success",
					Device: o.GetDevice(network),
				}
				if rxc, err := ParseInt64(columns[2]); err == nil {
					client.RxBytes = rxc
				}
				if txc, err := ParseInt64(columns[3]); err == nil {
					client.TxBytes = txc
				}
				if uptime, err := time.Parse(time.ANSIC, columns[4]); err == nil {
					client.Uptime = uptime.Unix()
					client.AliveTime = time.Now().Unix() - client.Uptime

				}
				clients[name] = client
			}
		case "routing":
			if len(columns) == 4 {
				name := columns[1]
				address := columns[0]
				if client, ok := clients[name]; ok {
					client.Address = address
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return clients, nil
}

func (o *_ovClient) statusFile(name string) string {
	return filepath.Join(o.Directory, name, "server.status")
}

func (o *_ovClient) readStatus(network string) map[string]*schema.OvClient {
	reader, err := os.Open(o.statusFile(network))
	if err != nil {
		libol.Debug("_ovClient.readStatus %v", err)
		return nil
	}
	defer reader.Close()
	if clients, err := o.scanStatus(network, reader); err != nil {
		libol.Warn("_ovClient.readStatus %v", err)
		return nil
	} else {
		return clients
	}
}

func (o *_ovClient) List(name string) <-chan *schema.OvClient {
	c := make(chan *schema.OvClient, 128)

	clients := o.readStatus(name)
	go func() {
		for _, v := range clients {
			c <- v
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}

var OvClient = _ovClient{
	Directory: config.VarDir("openvpn"),
}
