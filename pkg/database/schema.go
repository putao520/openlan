package database

type Switch struct {
	UUID            string            `ovsdb:"_uuid"`
	Protocol        string            `ovsdb:"protocol"`
	Listen          int               `ovsdb:"listen"`
	OtherConfig     map[string]string `ovsdb:"other_config" yaml:"other_config"`
	VirtualNetworks []string          `ovsdb:"virtual_networks" yaml:"virtual_networks"`
}

type VirtualNetwork struct {
	UUID         string            `ovsdb:"_uuid"`
	Name         string            `ovsdb:"name"`
	Bridge       string            `ovsdb:"bridge"`
	Address      string            `ovsdb:"address"`
	OtherConfig  map[string]string `ovsdb:"other_config" yaml:"other_config"`
	RemoteLinks  []string          `ovsdb:"remote_links" yaml:"remote_links"`
	LocalLinks   []string          `ovsdb:"local_links" yaml:"local_links"`
	OpenVPN      []string          `ovsdb:"openvpn" yaml:"openvpn"`
	PrefixRoutes []string          `ovsdb:"prefix_routes" yaml:"prefix_routes"`
}
