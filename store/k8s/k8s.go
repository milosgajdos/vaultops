package k8s

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// K8s implements kubernetes store
// It stores the secrets in kubernetes secret
// under provided secret key
type K8s struct {
	client kubernetes.Interface
	secret string
	key    string
}

// NewStore creates a new Kubernetes secret store handle and returns it.
// It returns error if the kubernetes client fails to be initialized.
func NewStore(secret, key string) (*K8s, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) == 0 {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure k8s client: %s", err.Error())
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed building k8s clientset: %s", err.Error())
	}

	return &K8s{
		client: client,
		secret: secret,
		key:    key,
	}, fmt.Errorf("not implemented")
}

// Write writes data to K8s store
func (k *K8s) Write(b []byte) (n int, err error) {
	return 0, fmt.Errorf("not implemented")
}

// Read reads data from k8s store
func (k *K8s) Read(b []byte) (n int, err error) {
	return 0, fmt.Errorf("not implemented")
}
