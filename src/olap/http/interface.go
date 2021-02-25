package http

import "github.com/danieldin95/openlan-go/src/config"

type Pointer interface {
	UUID() string
	Config() *config.Point
}
