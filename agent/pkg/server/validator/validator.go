package validator

import (
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

const (
	defaultMinReplicas     = 0
	defaultMaxReplicas     = 1
	maxReplicas            = 5
	defaultTargetLoad      = 100
	defaultZeroDuration    = 300
	defaultStartupDuration = 600
	defaultBuildDuration   = "40m"
	defaultHTTPProbePath   = "/"
)

var (
	dnsValidRegex = `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
)

type Validator struct {
	validDNS *regexp.Regexp
}

func New() *Validator {
	return &Validator{
		validDNS: regexp.MustCompile(dnsValidRegex),
	}
}

// Validates that the service name is valid for Kubernetes
func (v Validator) ValidateService(service string) error {
	matched := v.validDNS.MatchString(service)
	if matched {
		return nil
	}

	return fmt.Errorf("service: (%s) is invalid, must be a valid DNS entry", service)
}

// DefaultDeployRequest sets default values for the deploy request.
func (v Validator) DefaultDeployRequest(request *types.InferenceDeployment) {
	if request.Spec.Scaling == nil {
		request.Spec.Scaling = &types.ScalingConfig{}
	}

	if request.Spec.Scaling.MinReplicas == nil {
		request.Spec.Scaling.MinReplicas = new(int32)
		*request.Spec.Scaling.MinReplicas = defaultMinReplicas
	}

	if request.Spec.Scaling.MaxReplicas == nil {
		request.Spec.Scaling.MaxReplicas = new(int32)
		*request.Spec.Scaling.MaxReplicas = defaultMinReplicas
	}

	if request.Spec.Scaling.TargetLoad == nil {
		request.Spec.Scaling.TargetLoad = new(int32)
		*request.Spec.Scaling.TargetLoad = defaultTargetLoad
	}

	if request.Spec.Scaling.Type == nil {
		request.Spec.Scaling.Type = new(types.ScalingType)
		*request.Spec.Scaling.Type = types.ScalingTypeCapacity
	}

	if request.Spec.Scaling.ZeroDuration == nil {
		request.Spec.Scaling.ZeroDuration = new(int32)
		*request.Spec.Scaling.ZeroDuration = defaultZeroDuration
	}

	if request.Spec.Scaling.StartupDuration == nil {
		request.Spec.Scaling.StartupDuration = new(int32)
		*request.Spec.Scaling.StartupDuration = defaultStartupDuration
	}

	if request.Spec.Framework == "" {
		request.Spec.Framework = types.FrameworkOther
	}
}

// ValidateDeployRequest validates that the service name is valid for Kubernetes
func (v Validator) ValidateDeployRequest(request *types.InferenceDeployment) error {

	if request.Spec.Name == "" {
		return fmt.Errorf("service: is required")
	}

	err := v.ValidateService(request.Spec.Name)
	if err != nil {
		return err
	}

	if request.Spec.Image == "" {
		return fmt.Errorf("image: is required")
	}

	if request.Spec.Scaling == nil {
		return fmt.Errorf("scaling: is required")
	}

	if request.Spec.Framework == types.FrameworkOther {
		if request.Spec.Port == nil {
			return fmt.Errorf("port: is required for other framework")
		}
	}

	return nil
}

func (v Validator) ValidateBuildRequest(request *types.Build) error {
	if request.Spec.Name == "" {
		return fmt.Errorf("name: is required")
	}

	if request.Spec.Repository == "" {
		return fmt.Errorf("git repository: is required")
	}

	if request.Spec.ArtifactImage == "" {
		return fmt.Errorf("artifact image: is required")
	}

	if request.Spec.ProjectID == "" {
		return fmt.Errorf("project id: is required")
	}

	return nil
}

func (v Validator) DefaultBuildRequest(request *types.Build) {
	if request.Spec.Builder == "" {
		request.Spec.Builder = types.BuilderTypeDockerfile
	}
	if request.Spec.Branch == "" && request.Spec.Revision == "" {
		request.Spec.Branch = "main"
	}
	if request.Spec.ArtifactImageTag == "" {
		request.Spec.ArtifactImageTag = rand.String(8)
	}

	if request.Spec.Duration == "" {
		request.Spec.Duration = defaultBuildDuration
	}
}
