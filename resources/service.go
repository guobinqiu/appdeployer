package resources

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type ServiceOptions struct {
	ApplicationName string
	Namespace       string
	Port            int32
	TargetPort      int32
}

func CreateOrUpdateService(clientset *kubernetes.Clientset, ctx context.Context, opts ServiceOptions) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.ApplicationName,
			Namespace: opts.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       opts.Port,
					TargetPort: intstr.FromInt32(opts.TargetPort),
				},
			},
			Selector: map[string]string{
				"name": opts.ApplicationName,
			},
		},
	}

	if _, err := clientset.CoreV1().Services(opts.Namespace).Create(ctx, service, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}
