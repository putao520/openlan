package config

import "github.com/danieldin95/openlan-go/vswitch/schema"

func NewVersionSchema() schema.Version {
	return schema.Version{
		Version: Version,
		Date:    Date,
		Commit:  Commit,
	}
}
