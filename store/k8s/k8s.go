package k8s

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	DefaultTimeout = 5 * time.Second
)

// K8s implements kubernetes store
// It stores the secrets in kubernetes secret
// under provided secret key
type K8s struct {
	client kubernetes.Interface
	secret string
	key    string
	ns     string
	reader *bytes.Buffer
	ready  bool
}

// NewStore creates a new Kubernetes secret store handle and returns it.
// It returns error if the kubernetes client fails to be initialized.
func NewStore(secret, key, ns string) (*K8s, error) {
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
		ns:     ns,
		ready:  false,
	}, nil
}

// Write writes data to K8s store
func (k *K8s) Write(b []byte) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	secret, err := k.client.CoreV1().Secrets(k.ns).Get(ctx, k.secret, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		s := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: k.secret,
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				k.key: b,
			},
		}

		if _, err := k.client.CoreV1().Secrets(k.ns).Create(ctx, s, metav1.CreateOptions{}); err != nil {
			return 0, fmt.Errorf("failed to create secret %s in namespace %s: %v", k.secret, k.ns, err)
		}

		return len(b), nil
	}

	// compare the bytes and only update existing secrets if the bytes are not the same
	if !bytes.Equal(secret.Data[k.key], b) {
		secret.Data[k.key] = b

		if _, err := k.client.CoreV1().Secrets(k.ns).Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
			return 0, fmt.Errorf("failed to update secret %s in namespace %s: %v", k.secret, k.ns, err)
		}
	}

	return len(b), nil
}

// Read reads data from k8s store
func (k *K8s) Read(b []byte) (int, error) {
	if !k.ready {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
		defer cancel()

		secret, err := k.client.CoreV1().Secrets(k.ns).Get(ctx, k.secret, metav1.GetOptions{})

		if errors.IsNotFound(err) {
			return 0, fmt.Errorf("failed to find secret %s in namespace %s: %v", k.secret, k.ns, err)
		}

		if err != nil {
			return 0, fmt.Errorf("failed to read secret %s in namespace %s: %v", k.secret, k.ns, err)
		}

		k.reader = bytes.NewBuffer(secret.Data[k.key])
		k.ready = true
	}

	n, err := k.reader.Read(b)
	if err != nil {
		k.ready = false
	}

	return n, err
}
