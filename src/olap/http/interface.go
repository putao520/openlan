package http

import "github.com/danieldin95/openlan/src/config"

type Pointer interface {
	UUID() string
	Config() *config.Point
}
