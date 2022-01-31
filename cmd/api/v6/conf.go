package v6

import (
	"context"
	"github.com/danieldin95/openlan/cmd/api"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/go-logr/stdr"
	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"
	"log"
	"os"
)

var doing = context.Background()

type Conf struct {
	Server   string
	Database string
	OvS      client.Client
	Verbose  bool
}

func (c *Conf) Open() error {
	server := c.Server
	database := c.Database
	models := map[string]model.Model{
		"Global_Switch": &GlobalSwitch{},
	}
	dbModel, err := model.NewClientDBModel(database, models)
	if err != nil {
		return err
	}
	ops := []client.Option{
		client.WithEndpoint(server),
	}
	if !c.Verbose {
		// create a new logger to log to /dev/null
		writer, err := libol.OpenWrite(os.DevNull)
		if err != nil {
			writer = os.Stderr
		}
		l := stdr.NewWithOptions(log.New(writer, "", log.LstdFlags), stdr.Options{LogCaller: stdr.All})
		ops = append(ops, client.WithLogger(&l))
	}
	ovs, err := client.NewOVSDBClient(dbModel, ops...)
	if err != nil {
		return err
	}
	if err := ovs.Connect(doing); err != nil {
		return err
	}
	if _, err := ovs.MonitorAll(doing); err != nil {
		return err
	}
	c.OvS = ovs
	return nil
}

var conf *Conf

func GetConf() (*Conf, error) {
	var err error
	if conf == nil {
		conf = &Conf{
			Server:   api.Server,
			Database: api.Database,
			Verbose:  api.Verbose,
		}
		err = conf.Open()
	}
	return conf, err
}
