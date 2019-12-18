package api

import (
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/models"
)

type Worker interface {
	GetId() string
	GetServer() *libol.TcpServer
	NewTap() (*models.TapDevice, error)
	Send(*models.TapDevice, []byte)
}
