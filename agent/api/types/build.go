package types

type Build struct {
	Spec   BuildSpec   `json:"spec"`
	Status BuildStatus `json:"status,omitempty"`
}

type BuildSpec struct {
	Name                string `json:"name,omitempty"`
	ProjectID           string `json:"project_id,omitempty"`
	Namespace           string `json:"namespace,omitempty"`
	GitRepositorySource `json:",inline,omitempty"`
	BuildSource         `json:",inline,omitempty"`
}

type BuildSource struct {
	// directory is the target directory name.
	// Must not contain or start with '..'.  If '.' is supplied, the volume directory will be the
	// git repository.  Otherwise, if specified, the volume will contain the git repository in
	// the subdirectory with the given name.
	// +optional
	Directory string `json:"directory,omitempty"`

	Builder          BuilderType `json:"builder,omitempty"`
	ArtifactImage    string      `json:"image,omitempty"`
	ArtifactImageTag string      `json:"image_tag,omitempty"`

	Duration string `json:"duration,omitempty"`
}

type BuildStatus struct {
	Image string     `json:"image,omitempty"`
	Phase BuildPhase `json:"phase,omitempty"`
}

type BuildPhase string

const (
	BuildPhasePending   BuildPhase = "Pending"
	BuildPhaseRunning   BuildPhase = "Running"
	BuildPhaseSucceeded BuildPhase = "Succeeded"
	BuildPhaseFailed    BuildPhase = "Failed"
)

type BuilderType string

const (
	BuilderTypeDockerfile BuilderType = "Dockerfile"
	BuilderTypeENVD       BuilderType = "envd"
)

type GitRepositorySource struct {
	// repository is the URL
	Repository string `json:"repository"`
	Branch     string `json:"branch,omitempty"`
	// revision is the commit hash for the specified revision.
	// +optional
	Revision string `json:"revision,omitempty"`
}
