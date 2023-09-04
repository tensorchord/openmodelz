package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

func (cli *Client) RegisterAgent(ctx context.Context, token string, cluster types.ManagedCluster) (string, string, error) {
	urlValues := url.Values{}
	agentToken, err := ParseAgentToken(token)
	if err != nil {
		return "", "", err
	}
	urlPath := fmt.Sprintf(modelzCloudClusterWithUserControlPlanePath, agentToken.UID)
	headers := make(map[string][]string)
	headers["Authorization"] = []string{"Bearer " + agentToken.Token}

	resp, err := cli.post(ctx, urlPath, urlValues, cluster, headers)
	if err != nil {
		return "", "", wrapResponseError(err, resp, "register agent to modelz cloud", agentToken.UID)
	}
	defer ensureReaderClosed(resp)

	err = json.NewDecoder(resp.body).Decode(&cluster)
	if err != nil {
		return "", "", err
	}
	return cluster.ID, cluster.TokenID, nil
}

func (cli *Client) UpdateAgentStatus(ctx context.Context, token string, cluster types.ManagedCluster) error {
	urlValues := url.Values{}
	agentToken, err := ParseAgentToken(token)
	if err != nil {
		return err
	}
	urlPath := fmt.Sprintf(modelzCloudClusterControlPlanePath, agentToken.UID, cluster.ID)
	headers := make(map[string][]string)
	headers["Authorization"] = []string{"Bearer " + agentToken.Token}

	resp, err := cli.put(ctx, urlPath, urlValues, cluster, headers)
	if err != nil {
		return wrapResponseError(err, resp, "update agent status to modelz cloud", agentToken.UID)
	}
	defer ensureReaderClosed(resp)

	if resp.statusCode == 200 {
		return nil
	}
	return fmt.Errorf("failed to update agent status to modelz cloud, status code: %d", resp.statusCode)
}

func (cli *Client) GetAPIKeys(ctx context.Context, token string, cluster string) (types.APIKeyMap, error) {
	urlValues := url.Values{}
	agentToken, err := ParseAgentToken(token)
	keys := types.APIKeyMap{}
	if err != nil {
		return keys, err
	}
	headers := make(map[string][]string)
	headers["Authorization"] = []string{"Bearer " + agentToken.Token}

	urlPath := fmt.Sprintf(modelzCloudClusterAPIKeyControlPlanePath, agentToken.UID, cluster)
	resp, err := cli.get(ctx, urlPath, urlValues, headers)
	if err != nil {
		return keys, wrapResponseError(err, resp, "get api keys from modelz cloud", agentToken.UID)
	}
	defer ensureReaderClosed(resp)

	err = json.NewDecoder(resp.body).Decode(&keys)
	if err != nil {
		return keys, err
	}
	return keys, nil
}

func (cli *Client) GetNamespaces(ctx context.Context, token string, cluster string) (types.NamespaceList, error) {
	urlValues := url.Values{}
	agentToken, err := ParseAgentToken(token)
	ns := types.NamespaceList{}
	if err != nil {
		return ns, err
	}
	urlValues.Add("login_name", agentToken.UID)
	headers := make(map[string][]string)
	headers["Authorization"] = []string{"Bearer " + agentToken.Token}

	resp, err := cli.get(ctx, fmt.Sprintf(modelzCloudClusterNamespaceControlPlanePath, agentToken.UID, cluster), urlValues, headers)
	if err != nil {
		return ns, wrapResponseError(err, resp, "get namespaces from modelz cloud", agentToken.UID)
	}
	defer ensureReaderClosed(resp)

	err = json.NewDecoder(resp.body).Decode(&ns)
	if err != nil {
		return ns, err
	}

	ns.Items = append(ns.Items, GetNamespaceByUserID(agentToken.UID))
	return ns, nil
}

func (cli *Client) GetUIDFromDeploymentID(ctx context.Context, token string, cluster string, deployment string) (string, error) {
	urlValues := url.Values{}
	agentToken, err := ParseAgentToken(token)
	if err != nil {
		return "", err
	}
	headers := make(map[string][]string)
	headers["Authorization"] = []string{"Bearer " + agentToken.Token}
	urlPath := fmt.Sprintf(modelzCloudClusterDeploymentControlPlanePath, agentToken.UID, cluster, deployment)

	resp, err := cli.get(ctx, urlPath, urlValues, headers)
	if err != nil {
		return "", err
	}
	defer ensureReaderClosed(resp)

	var uid string
	err = json.NewDecoder(resp.body).Decode(&uid)
	if err != nil {
		return "", err
	}

	if resp.statusCode == 200 {
		return uid, nil
	}
	return "", fmt.Errorf("failed to get uid from deployment id, status code: %d", resp.statusCode)
}
