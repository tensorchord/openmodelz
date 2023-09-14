package types

import "time"

const (
	ClusterStatusInit    = "init"
	ClusterStatusActive  = "active"
	ClusterStatusUnknown = "unknown"
)

const (
	DailEndPointSuffix = "/api/v1/clusteragent/connect"
)

type AgentToken struct {
	UID   string `json:"uid,omitempty"`
	Token string `json:"token,omitempty"`
	Type  string `json:"type,omitempty"`
}

type ManagedCluster struct {
	ID                string    `json:"id,omitempty"`
	TokenID           string    `json:"token_id,omitempty"`
	Version           string    `json:"version,omitempty"`
	KubernetesVersion string    `json:"kubernetes_version,omitempty"`
	Platform          string    `json:"platform,omitempty"`
	Status            string    `json:"status,omitempty"`
	CreatedAt         time.Time `json:"created_at,omitempty"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
	Region            string    `json:"region,omitempty"`
	ServerResources   string    `json:"server_resources,omitempty"`
}

type APIKeyMap map[string]string

type NamespaceList struct {
	Items []string `json:"items,omitempty"`
}
