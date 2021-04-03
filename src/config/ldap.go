package config

type LDAP struct {
	Server       string `json:"server"`
	BindDN       string `json:"bindDN"`
	Password     string `json:"password"`
	BaseDN       string `json:"baseDN"`
	Attribute    string `json:"attribute"`
	SearchFilter string `json:"searchFilter"`
	EnableTls    bool   `json:"enableTls"`
}
