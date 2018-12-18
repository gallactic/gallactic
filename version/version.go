package version

// Version components
const (
	Maj = "0"
	Min = "4"
	Fix = "0"
)

var (
	// Version is the current version of Gallactic in string
	Version = "0.4.0"

	// GitCommit is the current HEAD set using ldflags.
	GitCommit string
)

func init() {
	if GitCommit != "" {
		Version += "-" + GitCommit
	}
}
