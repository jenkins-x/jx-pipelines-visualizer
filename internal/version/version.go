package version

// these are set at compile time by GoReleaser through LD Flags
var (
	Version  = "dev"
	Revision = "unknown"
	Date     = "now"
)
