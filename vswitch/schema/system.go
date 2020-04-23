package schema

type Version struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Commit  string `json:"commit"`
}

type Worker struct {
	Uptime int64  `json:"uptime"`
	UUID   string `json:"uuid"`
	Alias  string `json:"alias"`
}
