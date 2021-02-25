package libol

var (
	Date    string
	Version string
	Commit  string
)

func init() {
	Debug("libol: version is %s", Version)
	Debug("libol: built on %s", Date)
	Debug("libol: commit at %s", Commit)
}
