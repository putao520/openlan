package schema

type ACL struct {
	Name  string    `json:"name"`
	Rules []ACLRule `json:"rules"`
}

type ACLRule struct {
	Name    string `json:"name"`
	IpSrc   string `json:"source"`
	IpDst   string `json:"dest"`
	SrcPort string `json:"sourcePort"`
	DstPort string `json:"destPort"`
	Action  string `json:"action"`
}
