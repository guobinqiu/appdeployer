package resources

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceAccountOptions struct {
	ApplicationName string
	Namespace       string
}

func CreateOrUpdateServiceAccount(clientset *kubernetes.Clientset, ctx context.Context, opts ServiceAccountOptions) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.ApplicationName,
			Namespace: opts.Namespace,
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: "docker-" + opts.ApplicationName,
			},
		},
	}

	if _, err := clientset.CoreV1().ServiceAccounts(opts.Namespace).Create(ctx, serviceAccount, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	fmt.Println("kube serviceaccount successfully done.")

	return nil
}
