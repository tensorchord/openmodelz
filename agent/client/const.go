// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

const DefaultModelzGatewayHost = "http://0.0.0.0:8080"

const defaultProto = "http"
const defaultAddr = "0.0.0.0:8080"

// Base path for api, distinguish from frontend pages
const apiBasePath = ""

const (
	gatewayInferControlPlanePath                 = "/system/inferences"
	gatewayInferScaleControlPath                 = "/system/scale-inference"
	gatewayInferInstanceControlPlanePath         = "/system/inference/%s/instances"
	gatewayInferInstanceExecControlPlanePath     = "/system/inference/%s/instance/%s/exec"
	gatewayServerControlPlanePath                = "/system/servers"
	gatewayServerLabelCreateControlPlanePath     = "/system/server/%s/labels"
	gatewayServerNodeDeleteControlPlanePath      = "/system/server/%s/delete"
	gatewayNamespaceControlPlanePath             = "/system/namespaces"
	gatewayBuildControlPlanePath                 = "/system/build"
	gatewayBuildInstanceControlPlanePath         = "/system/build/%s"
	gatewayImageCacheControlPlanePath            = "/system/image-cache"
	modelzCloudClusterControlPlanePath           = "/api/v1/users/%s/clusters/%s"
	modelzCloudClusterWithUserControlPlanePath   = "/api/v1/users/%s/clusters"
	modelzCloudClusterAPIKeyControlPlanePath     = "/api/v1/users/%s/clusters/%s/api_keys"
	modelzCloudClusterNamespaceControlPlanePath  = "/api/v1/users/%s/clusters/%s/namespaces"
	modelzCloudClusterDeploymentControlPlanePath = "/api/v1/users/%s/clusters/%s/deployments/%s/agent"
)

const (
	// EnvOverrideHost is the name of the environment variable that can be used
	// to override the default host to connect to (DefaultEnvdServerHost).
	//
	// This env-var is read by FromEnv and WithHostFromEnv and when set to a
	// non-empty value, takes precedence over the default host (which is platform
	// specific), or any host already set.
	EnvOverrideHost = "MODELZ_GATEWAY_HOST"

	// EnvOverrideCertPath is the name of the environment variable that can be
	// used to specify the directory from which to load the TLS certificates
	// (ca.pem, cert.pem, key.pem) from. These certificates are used to configure
	// the Client for a TCP connection protected by TLS client authentication.
	//
	// TLS certificate verification is enabled by default if the Client is configured
	// to use a TLS connection. Refer to EnvTLSVerify below to learn how to
	// disable verification for testing purposes.
	//
	//
	// For local access to the API, it is recommended to connect with the daemon
	// using the default local socket connection (on Linux), or the named pipe
	// (on Windows).
	//
	// If you need to access the API of a remote daemon, consider using an SSH
	// (ssh://) connection, which is easier to set up, and requires no additional
	// configuration if the host is accessible using ssh.
	EnvOverrideCertPath = "ENVD_SERVER_CERT_PATH"

	// EnvTLSVerify is the name of the environment variable that can be used to
	// enable or disable TLS certificate verification. When set to a non-empty
	// value, TLS certificate verification is enabled, and the client is configured
	// to use a TLS connection, using certificates from the default directories
	// (within `~/.envd`); refer to EnvOverrideCertPath above for additional
	// details.
	//
	//
	// Before setting up your client and daemon to use a TCP connection with TLS
	// client authentication, consider using one of the alternatives mentioned
	// in EnvOverrideCertPath above.
	//
	// Disabling TLS certificate verification (for testing purposes)
	//
	// TLS certificate verification is enabled by default if the Client is configured
	// to use a TLS connection, and it is highly recommended to keep verification
	// enabled to prevent machine-in-the-middle attacks.
	//
	// Set the "ENVD_SERVER_TLS_VERIFY" environment to an empty string ("") to
	// disable TLS certificate verification. Disabling verification is insecure,
	// so should only be done for testing purposes. From the Go documentation
	// (https://pkg.go.dev/crypto/tls#Config):
	//
	// InsecureSkipVerify controls whether a client verifies the server's
	// certificate chain and host name. If InsecureSkipVerify is true, crypto/tls
	// accepts any certificate presented by the server and any host name in that
	// certificate. In this mode, TLS is susceptible to machine-in-the-middle
	// attacks unless custom verification is used. This should be used only for
	// testing or in combination with VerifyConnection or VerifyPeerCertificate.
	EnvTLSVerify = "ENVD_SERVER_TLS_VERIFY"
)
