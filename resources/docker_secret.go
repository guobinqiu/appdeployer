package resources

import (
	"context"

	"github.com/guobinqiu/deployer/docker"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DockerSecretOptions struct {
	ApplicationName string
	Namespace       string
	docker.DockerOptions
}

func CreateOrUpdateDockerSecret(clientset *kubernetes.Clientset, ctx context.Context, opts DockerSecretOptions) error {
	dockerconfigjson, err := docker.BuildDockerAuthConfig(opts.DockerOptions)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "docker-" + opts.ApplicationName,
			Namespace: opts.Namespace,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": dockerconfigjson,
		},
	}

	if _, err = clientset.CoreV1().Secrets(opts.Namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}
