package api

import (
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/vswitch/models"
	"github.com/songgao/water"
)

type Worker interface {
	GetRedis() *libol.RedisCli
	GetServer() *libol.TcpServer
	GetUser(name string) *models.User
	NewTap() (*water.Interface, error)
	AddPoint(p *models.Point)
}
