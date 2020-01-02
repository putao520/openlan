package api

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
)

type Worker interface {
	GetId() string
	GetServer() *libol.TcpServer
	NewTap() (models.Taper, error)
	Write(models.Taper, []byte)
}