package k8s

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tensorchord/openmodelz/agent/api/types"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	secretsMountPath             = "/var/modelz/secrets"
	secretLabel                  = "app.kubernetes.io/managed-by"
	secretLabelValue             = "modelz"
	secretsProjectVolumeNameTmpl = "projected-secrets"
)

// SecretsClient exposes the standardized CRUD behaviors for Kubernetes secrets.  These methods
// will ensure that the secrets are structured and labelled correctly for use by the modelz system.
type SecretsClient interface {
	// List returns a list of available function secrets.  Only the names are returned
	// to ensure we do not accidentally read or print the sensitive values during
	// read operations.
	List(namespace string) (names []string, err error)
	// Create adds a new secret, with the appropriate labels and structure to be
	// used as a function secret.
	Create(secret types.Secret) error
	// Replace updates the value of a function secret
	Replace(secret types.Secret) error
	// Delete removes a function secret
	Delete(name string, namespace string) error
	// GetSecrets queries Kubernetes for a list of secrets by name in the given k8s namespace.
	// This should only be used if you need access to the actual secret structure/value. Specifically,
	// inside the FunctionFactory.
	GetSecrets(namespace string, secretNames []string) (map[string]*apiv1.Secret, error)
}

// SecretInterfacer exposes the SecretInterface getter for the k8s client.
// This is implemented by the CoreV1Interface() interface in the Kubernetes client.
// The SecretsClient only needs this one interface, but needs to be able to set the
// namespaces when the interface is instantiated, meaning, we need the Getter and not the
// SecretInterface itself.
type SecretInterfacer interface {
	// Secrets returns a SecretInterface scoped to the specified namespace
	Secrets(namespace string) typedv1.SecretInterface
}

type SecretClient struct {
	kube SecretInterfacer
}

// NewSecretsClient constructs a new SecretsClient using the provided Kubernetes client.
func NewSecretClient(kube kubernetes.Interface) SecretsClient {
	return &SecretClient{
		kube: kube.CoreV1(),
	}
}

func (c SecretClient) List(namespace string) (names []string, err error) {
	res, err := c.kube.Secrets(namespace).List(context.TODO(), c.selector())
	if err != nil {
		log.Printf("failed to list secrets in %s: %v\n", namespace, err)
		return nil, err
	}

	names = make([]string, len(res.Items))
	for idx, item := range res.Items {
		// this is safe because size of names matches res.Items exactly
		names[idx] = item.Name
	}
	return names, nil
}

func (c SecretClient) Create(secret types.Secret) error {
	err := c.validateSecret(secret)
	if err != nil {
		return err
	}

	req := &apiv1.Secret{
		Type: apiv1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Labels: map[string]string{
				secretLabel: secretLabelValue,
			},
		},
	}

	if len(secret.Data) > 0 {
		req.Data = secret.Data
	}

	if len(secret.StringData) > 0 {
		req.StringData = secret.StringData
	}

	s, err := c.kube.Secrets(secret.Namespace).Get(context.Background(), secret.Name, metav1.GetOptions{})
	if err == nil && s != nil {
		log.Printf("secret %s.%s already exists\n", secret.Name, secret.Namespace)
		return nil
	}

	_, err = c.kube.Secrets(secret.Namespace).Create(context.TODO(), req, metav1.CreateOptions{})
	if err != nil {
		log.Printf("failed to create secret %s.%s: %v\n", secret.Name, secret.Namespace, err)
		return err
	}

	log.Printf("created secret %s.%s\n", secret.Name, secret.Namespace)

	return nil
}

func (c SecretClient) Replace(secret types.Secret) error {
	err := c.validateSecret(secret)
	if err != nil {
		return err
	}

	kube := c.kube.Secrets(secret.Namespace)
	found, err := kube.Get(context.TODO(), secret.Name, metav1.GetOptions{})
	if err != nil {
		log.Printf("can not retrieve secret for update %s.%s: %v\n", secret.Name, secret.Namespace, err)
		return err
	}

	_, err = kube.Update(context.TODO(), found, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("can not update secret %s.%s: %v\n", secret.Name, secret.Namespace, err)
		return err
	}

	return nil
}

func (c SecretClient) Delete(namespace string, name string) error {
	err := c.kube.Secrets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("can not delete %s.%s: %v\n", name, namespace, err)
	}
	return err
}

func (c SecretClient) GetSecrets(namespace string, secretNames []string) (map[string]*apiv1.Secret, error) {
	kube := c.kube.Secrets(namespace)
	opts := metav1.GetOptions{}

	secrets := map[string]*apiv1.Secret{}
	for _, secretName := range secretNames {
		secret, err := kube.Get(context.TODO(), secretName, opts)
		if err != nil {
			return nil, err
		}
		secrets[secretName] = secret
	}

	return secrets, nil
}

func (c SecretClient) selector() metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", secretLabel, secretLabelValue),
	}
}

func (c SecretClient) validateSecret(secret types.Secret) error {
	if strings.TrimSpace(secret.Namespace) == "" {
		return errors.New("namespace may not be empty")
	}

	if strings.TrimSpace(secret.Name) == "" {
		return errors.New("name may not be empty")
	}

	return nil
}
