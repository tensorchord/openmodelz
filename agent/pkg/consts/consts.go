package consts

import "time"

const (
	LabelBuildName            = "ai.tensorchord.build"
	LabelName                 = "ai.tensorchord.name"
	LabelServerResource       = "ai.tensorchord.server-resource"
	AnnotationControlPlaneKey = "ai.tensorchord.control-plane"
	ModelzAnnotationValue     = "modelz"

	Domain        = "modelz.live"
	DefaultPrefix = "modelz-"
	APIKEY_PREFIX = "mzi-"
)
const DefaultAPIServerReadyTimeout = 15 * time.Minute
