package version

// These variables are set at build time using ldflags:
//
//	go build -ldflags="-X repoctr/internal/version.Version=v1.0.0"
var (
	// Version is the current version of repo-ctr (set via ldflags)
	Version = "dev"

	// GitHubOwner is the GitHub repository owner
	GitHubOwner = "aykutkilic"

	// GitHubRepo is the GitHub repository name
	GitHubRepo = "repoctr"
)
