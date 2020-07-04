package libol

var (
	Date    string
	Version string
	Commit  string
)

func init() {
	Info("libol: version is %s", Version)
	Info("libol: built on %s", Date)
	Info("libol: commit at %s", Commit)
}
