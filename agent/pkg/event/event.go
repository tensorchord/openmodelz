package event

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/client"
)

type Interface interface {
	CreateDeploymentEvent(namespace, deployment, event, message string) error
}

type EventRecorder struct {
	Client     *client.Client
	AgentToken string
}

func NewEventRecorder(client *client.Client, token string) Interface {
	return &EventRecorder{
		Client:     client,
		AgentToken: token,
	}
}

func (e *EventRecorder) CreateDeploymentEvent(namespace, deployment, event, message string) error {
	user, err := client.GetUserIDFromNamespace(namespace)
	if err != nil {
		return err
	} else if user == "" {
		return fmt.Errorf("user id is empty")
	}

	deploymentEvent := types.DeploymentEvent{
		UserID:       user,
		DeploymentID: deployment,
		EventType:    event,
		Message:      message,
	}
	err = e.Client.CreateDeploymentEvent(context.TODO(), e.AgentToken, deploymentEvent)
	if err != nil {
		logrus.Errorf("failed to create deployment event: %v", err)
		return err
	}

	return nil
}
