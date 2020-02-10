package vswitch


type WorkerSchema struct {
	Uptime int64
	UUID   string
}

type PointSchema struct {
	Uptime  int64
	UUID    string
	Alias   string
	Address string
	Device  string
	RxBytes uint64
	TxBytes uint64
	ErrPkt  uint64
	State   string
}

type LinkSchema struct {
	Uptime  int64
	UUID    string
	Alias   string
	Address string
	IpAddr  string
	Device  string
	RxBytes uint64
	TxBytes uint64
	ErrPkt  uint64
	State   string
}

type NeighborSchema struct {
	Uptime int64
	UUID   string
	HwAddr string
	IpAddr string
	Client string
}

type OnLineSchema struct {
	EthType    uint16
	IpSource   string
	IpDest     string
	IpProto    string
	PortSource uint16
	PortDest   uint16
}

type IndexSchema struct {
	Worker    WorkerSchema
	Points    []PointSchema
	Links     []LinkSchema
	Neighbors []NeighborSchema
	OnLines   []OnLineSchema
}