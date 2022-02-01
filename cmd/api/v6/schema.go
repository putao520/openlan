package v6

type SwitchDB struct {
	UUID        string            `ovsdb:"_uuid"`
	Protocol    string            `ovsdb:"protocol"`
	Listen      int               `ovsdb:"listen"`
	OtherConfig map[string]string `ovsdb:"other_config" yaml:"other_config"`
}

type NetworkDB struct {
	UUID        string            `ovsdb:"_uuid"`
	Name        string            `ovsdb:"name"`
	Bridge      string            `ovsdb:"bridge"`
	Address     string            `ovsdb:"address"`
	OtherConfig map[string]string `ovsdb:"other_config" yaml:"other_config"`
}
