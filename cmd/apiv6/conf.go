package apiv6

import (
	"context"
	"github.com/go-logr/stdr"
	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"
)

type Conf struct {
	Server   string
	Database string
	OvS      client.Client
	Verbose  bool
}

func (c *Conf) Open() error {
	server := c.Server
	database := c.Database
	dbModel, err := model.NewClientDBModel(database, map[string]model.Model{
		"Global_Switch": &GlobalSwitch{},
	})
	if err != nil {
		return err
	}
	stdr.SetVerbosity(0)
	ovs, err := client.NewOVSDBClient(dbModel,
		client.WithEndpoint(server),
		client.WithLogger(nil))
	if err != nil {
		return err
	}
	if !c.Verbose {
		stdr.SetVerbosity(0)
	}
	if err := ovs.Connect(context.Background()); err != nil {
		return err
	}
	if _, err := ovs.MonitorAll(context.Background()); err != nil {
		return err
	}
	c.OvS = ovs
	return nil
}

var conf *Conf

func NewConf(server, database string, Verbose bool) error {
	conf = &Conf{
		Server:   server,
		Database: database,
		Verbose:  Verbose,
	}
	return conf.Open()
}