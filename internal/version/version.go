package version

// Dynamically inserted at build time by goreleaser. See `ldflags` in .goreleaser.yml.
var (
	Version = "unknown"
	Commit  = "unknown"
	Date    = "unknown"
)
