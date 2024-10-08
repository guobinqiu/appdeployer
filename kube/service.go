package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type ServiceOptions struct {
	Name       string
	Namespace  string
	Port       int32 `form:"port" json:"port"`
	TargetPort int32
}

func CreateOrUpdateService(clientset *kubernetes.Clientset, ctx context.Context, opts ServiceOptions, logHandler func(msg string)) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "app",
					Port:       opts.Port,
					TargetPort: intstr.FromInt32(opts.TargetPort),
				},
			},
			Selector: map[string]string{
				"name": opts.Name,
			},
		},
	}

	if _, err := clientset.CoreV1().Services(opts.Namespace).Create(ctx, service, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create service resource: %v", err)
		}
		logHandler("service resource successfully updated")
	} else {
		logHandler("service resource successfully created")
	}

	return nil
}
