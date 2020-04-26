package http

import "github.com/danieldin95/openlan-go/main/config"

type Pointer interface {
	UUID   ()string
	Config () *config.Point
}
