package models

import (
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/schema"
)

type Link struct {
	User       string
	Network    string
	Protocol   string
	StatusFile string
}

func (l *Link) reload() *schema.Point {
	status := &schema.Point{}
	_ = libol.UnmarshalLoad(status, l.StatusFile)
	return status
}

func (l *Link) Status() *schema.Point {
	return l.reload()
}
