package version

import "fmt"

var (
	version = "1.0.0"
	commit  = "none"
	date    = "unknown"
)

type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

func Get() Info {
	return Info{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
}

func (i Info) String() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", i.Version, i.Commit, i.Date)
}

func FullVersion() string {
	return Get().String()
}
