package types

import "time"

const (
	DeploymentCreateEvent     = "deployment-create"
	DeploymentUpdateEvent     = "deployment-update"
	DeploymentDeleteEvent     = "deployment-delete"
	DeploymentScaleUpEvent    = "deployment-scale-up"
	DeploymentScaleDownEvent  = "deployment-scale-down"
	DeploymentScaleBlockEvent = "deployment-scale-block"
	PodCreateEvent            = "pod-create"
	PodReadyEvent             = "pod-ready"
	PodTimeoutEvent           = "pod-timeout"
)

type DeploymentEvent struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       string    `json:"user_id"`
	DeploymentID string    `json:"deployment_id"`
	EventType    string    `json:"event_type"`
	Message      string    `json:"message"`
}
