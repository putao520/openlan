package schema

import "github.com/danieldin95/openlan-go/src/libol"

type Version struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Commit  string `json:"commit"`
}

func NewVersionSchema() Version {
	return Version{
		Version: libol.Version,
		Date:    libol.Date,
		Commit:  libol.Commit,
	}
}
