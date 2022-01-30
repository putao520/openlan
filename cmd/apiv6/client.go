package apiv6

import (
	"context"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"
)

var ovs client.Client

func Open(server, database string) error {
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
	ovs = cli
	return nil
}

func List() {
	var lsList []GlobalSwitch
	if err := ovs.List(&lsList); err == nil {
		for _, ls := range lsList {
			libol.Debug("%s %d", ls.Protocol, ls.Listen)
		}
	}
	if len(lsList) == 0 {
		ops, _ := ovs.Create(&GlobalSwitch{
			Protocol: "tcp",
			Listen:   10002,
		})
		if ret, err := ovs.Transact(ops...); err != nil {
			libol.Error("%s", err)
		} else {
			libol.Debug("%s", ret)
		}
	}
}
