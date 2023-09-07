package server

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/client"
)

func (s *Server) initModelZCloud(url, token, region string) error {
	cluster := types.ManagedCluster{
		Region: region,
	}

	cli, err := client.NewClientWithOpts(
		client.WithHost(url))
	if err != nil {
		return errors.Wrap(err, "failed to create modelz cloud client")
	}
	s.modelzCloudClient = cli

	err = s.runtime.GetClusterInfo(&cluster)
	if err != nil {
		return errors.Wrap(err, "failed to get managed cluster info")
	}

	apiServerReady := make(chan struct{})
	go func() {
		if err := s.modelzCloudClient.WaitForAPIServerReady(); err != nil {
			logrus.Fatalf("failed to wait for apiserver ready: %v", err)
		}
		close(apiServerReady)
	}()

	cluster.Status = types.ClusterStatusInit
	// after init modelz cloud client, register agent
	clusterID, tokenID, err := cli.RegisterAgent(context.Background(), token, cluster)
	if err != nil {
		return errors.Wrap(err, "failed to register agent to modelz cloud")
	}
	s.config.ModelZCloud.ID = clusterID
	s.config.ModelZCloud.TokenID = tokenID

	apikeys, err := s.modelzCloudClient.GetAPIKeys(context.Background(), apiServerReady, s.config.ModelZCloud.AgentToken, s.config.ModelZCloud.ID)
	if err != nil {
		logrus.Errorf("failed to get apikeys: %v", err)
	}

	s.config.ModelZCloud.APIKeys = apikeys

	namespaces, err := s.modelzCloudClient.GetNamespaces(context.Background(), apiServerReady, s.config.ModelZCloud.AgentToken, s.config.ModelZCloud.ID)
	if err != nil {
		logrus.Errorf("failed to get namespaces: %v", err)
	}

	nss := []string{}
	for _, ns := range namespaces.Items {
		nss = append(nss, ns)
		err = s.runtime.NamespaceCreate(context.Background(), ns)
		if err != nil {
			logrus.Errorf("failed to create namespace %s: %v", ns, err)
			continue
		}
	}
	s.config.ModelZCloud.UserNamespaces = nss

	return nil
}
