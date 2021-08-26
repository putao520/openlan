package http

import "github.com/danieldin95/openlan/pkg/config"

type Pointer interface {
	UUID() string
	Config() *config.Point
}
