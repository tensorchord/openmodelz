package runtime

import "github.com/tensorchord/openmodelz/agent/api/types"

const (
	labelVendor = "ai.modelz.open.vendor"
	valueVendor = "openmodelz"

	labelName = "ai.modelz.open.name"
)

func expectedLabels(inf types.InferenceDeployment) map[string]string {
	return map[string]string{
		labelVendor: valueVendor,
		labelName:   inf.Spec.Name,
	}
}
