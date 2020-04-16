package schema

type Category struct {
	Name string `json:"name"`
}

//{
//"name": "YunStack",
//"value": 1,
//"symbolSize": 20,
//"category": 0,
//"id": 0,
//"label": {"show":  true
//}

type Label struct {
	Show bool `json:"show"`
}

type Node struct {
	Name       string `json:"name"`
	Value      int    `json:"value"`
	SymbolSize int    `json:"symbolSize"`
	Category   int    `json:"category"`
	Id         int    `json:"id"`
	Label      *Label `json:"label,omitempty"`
}

type Link struct {
	Source int `json:"source"`
	Target int `json:"target"`
	Weight int `json:"weight"`
}
