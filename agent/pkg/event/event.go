package event

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/tensorchord/openmodelz/agent/client"
	"github.com/tensorchord/openmodelz/agent/pkg/query"
)

type Interface interface {
	CreateDeploymentEvent(namespace, deployment, event, message string) error
}

type EventRecorder struct {
	DB query.Querier
}

func NewEventRecorder(q query.Querier) Interface {
	return &EventRecorder{
		DB: q,
	}
}

func (e *EventRecorder) CreateDeploymentEvent(namespace, deployment, event, message string) error {
	user, err := client.GetUserIDFromNamespace(namespace)
	if err != nil {
		return err
	} else if user == "" {
		return fmt.Errorf("user id is empty")
	}
	userId, err := uuid.Parse(user)
	if err != nil {
		return err
	}

	deploymentId, err := uuid.Parse(deployment)
	if err != nil {
		return err
	}

	params := query.CreateDeploymentEventParams{
		UserID:       uuid.NullUUID{UUID: userId, Valid: true},
		DeploymentID: uuid.NullUUID{UUID: deploymentId, Valid: true},
		EventType:    NullStringBuilder(event, true),
		Message:      NullStringBuilder(message, true),
	}
	if _, err := e.DB.CreateDeploymentEvent(context.TODO(), params); err != nil {
		return err
	}
	return nil
}
