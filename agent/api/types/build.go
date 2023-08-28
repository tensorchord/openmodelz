package types

type Build struct {
	Spec   BuildSpec   `json:"spec"`
	Status BuildStatus `json:"status,omitempty"`
}

type BuildSpec struct {
	Name                string `json:"name,omitempty"`
	Namespace           string `json:"namespace,omitempty"`
	GitRepositorySource `json:",inline,omitempty"`
	DockerSource        `json:",inline,omitempty"`
	BuildTarget         BuildTarget `json:",inline,omitempty"`
}

type DockerSource struct {
	ArtifactImage    string `json:"image,omitempty"`
	ArtifactImageTag string `json:"image_tag,omitempty"`
	AuthN            AuthN  `json:"authn,omitempty"`
	SecretID         string `json:"secret_id,omitempty"`
}

type BuildTarget struct {
	// directory is the target directory name.
	// Must not contain or start with '..'.  If '.' is supplied, the volume directory will be the
	// git repository.  Otherwise, if specified, the volume will contain the git repository in
	// the subdirectory with the given name.
	// +optional
	Directory string `json:"directory,omitempty"`

	Builder          BuilderType `json:"builder,omitempty"`
	ArtifactImage    string      `json:"image,omitempty"`
	ArtifactImageTag string      `json:"image_tag,omitempty"`
	Digest           string      `json:"digest,omitempty"`

	Duration      string `json:"duration,omitempty"`
	Registry      string `json:"registry,omitempty"`
	RegistryToken string `json:"registry_token,omitempty"`
}

type AuthN struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

type BuildStatus struct {
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
	BuilderTypeImage      BuilderType = "image"
)

type GitRepositorySource struct {
	// repository is the URL
	Repository string `json:"repository"`
	Branch     string `json:"branch,omitempty"`
	// revision is the commit hash for the specified revision.
	// +optional
	Revision string `json:"revision,omitempty"`
}
