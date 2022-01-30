package apiv6

import (
	"context"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"
)

type Client struct {
	Server   string
	Database string
	OvS      client.Client
}

func (c *Client) Open() error {
	server := c.Server
	database := c.Database
	dbModel, err := model.NewDBModel(database, map[string]model.Model{
		"Global_Switch": &GlobalSwitch{},
	})
	if err != nil {
		return err
	}
	cli, err := client.NewOVSDBClient(dbModel, client.WithEndpoint(server))
	if err != nil {
		return err
	}
	if err := cli.Connect(context.Background()); err != nil {
		return err
	}
	if _, err := cli.MonitorAll(); err != nil {
		return err
	}
	c.OvS = cli
	return nil
}

func (c *Client) List() {
	var lsList []GlobalSwitch
	if err := c.OvS.List(&lsList); err == nil {
		for _, ls := range lsList {
			libol.Debug("ovs.List %s %d", ls.Protocol, ls.Listen)
		}
	}
	if len(lsList) == 0 {
		ops, _ := c.OvS.Create(&GlobalSwitch{
			Protocol: "tcp",
			Listen:   10002,
		})
		if ret, err := c.OvS.Transact(ops...); err != nil {
			libol.Error("ovs.Transact %s", err)
		} else {
			libol.Debug("ovs.Transact %s", ret)
		}
	}
}

var ovs *Client

func NewOvS(server, database string) error {
	ovs = &Client{Server: server, Database: database}
	return ovs.Open()
}
