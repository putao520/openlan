package v6

type GlobalSwitch struct {
	UUID     string            `ovsdb:"_uuid"`
	Protocol string            `ovsdb:"protocol"`
	Listen   int               `ovsdb:"listen"`
	Config   map[string]string `ovsdb:"other_config"`
}
