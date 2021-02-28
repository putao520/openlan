package config

type ACL struct {
	Name  string     `json:"name"`
	Rules []*ACLRule `json:"rules"`
}

type ACLRule struct {
	Name    string `json:"name"`
	SrcIp   string `json:"source"`
	DstIp   string `json:"dest"`
	Proto   string `json:"proto"`
	SrcPort int    `json:"srcPort"`
	DstPort int    `json:"dstPort"`
	Action  string `json:"action"`
}

func (ru *ACLRule) Correct() {
}
