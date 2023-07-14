package types

// ProviderInfo provides information about the configured provider
type ProviderInfo struct {
	Name          string       `json:"provider"`
	Version       *VersionInfo `json:"version"`
	Orchestration string       `json:"orchestration"`
}

// VersionInfo provides the commit message, sha and release version number
type VersionInfo struct {
	Version      string `json:"version,omitempty"`
	BuildDate    string `json:"build_date,omitempty"`
	GitCommit    string `json:"git_commit,omitempty"`
	GitTag       string `json:"git_tag,omitempty"`
	GitTreeState string `json:"git_tree_state,omitempty"`
	GoVersion    string `json:"go_version,omitempty"`
	Compiler     string `json:"compiler,omitempty"`
	Platform     string `json:"platform,omitempty"`
}
