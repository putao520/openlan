package database

import (
	"github.com/ovn-org/libovsdb/model"
)

var models = map[string]model.Model{
	"Global_Switch":   &Switch{},
	"Virtual_Network": &VirtualNetwork{},
	"Virtual_Link":    &VirtualLink{},
	"Open_VPN":        &OpenVPN{},
}
