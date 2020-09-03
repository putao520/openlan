package ctrlc

import "github.com/danieldin95/openlan-go/src/libol"

type Storage struct {
	Point    *libol.SafeStrMap
	Link     *libol.SafeStrMap
	Neighbor *libol.SafeStrMap
	Switch   *libol.SafeStrMap
}

var Storager = Storage{
	Point:    libol.NewSafeStrMap(1024),
	Link:     libol.NewSafeStrMap(1024),
	Neighbor: libol.NewSafeStrMap(1024),
	Switch:   libol.NewSafeStrMap(1024),
}
