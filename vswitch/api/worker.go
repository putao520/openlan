package api

import (
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/models"
)

type Worker interface {
	GetId() string
	GetRedis() *libol.RedisCli
	GetServer() *libol.TcpServer
	GetUser(name string) *models.User
	NewTap() (*models.TapDevice, error)
	Send(*models.TapDevice, []byte)
}
