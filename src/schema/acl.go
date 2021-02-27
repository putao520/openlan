package schema

type ACL struct {
	Name  string    `json:"name"`
	Rules []ACLRule `json:"rules"`
}

type ACLRule struct {
	Name    string `json:"name"`
	SrcIp   string `json:"source"`
	DstIp   string `json:"dest"`
	Proto   string `json:"proto"`
	SrcPort int    `json:"sourcePort"`
	DstPort int    `json:"destPort"`
	Action  string `json:"action"`
}
