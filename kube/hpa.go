package kube

import (
	"context"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type HPAOptions struct {
	Name        string
	Namespace   string
	Enabled     bool  `form:"enabled" json:"enabled"`
	MinReplicas int32 `form:"minreplicas" json:"minreplicas"`
	MaxReplicas int32 `form:"maxreplicas" json:"maxreplicas"`
	CPURate     int32 `form:"cpurate" json:"cpurate"`
}

func CreateOrUpdateHPA(clientset *kubernetes.Clientset, ctx context.Context, opts HPAOptions, logHandler func(msg string)) error {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.Namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "app/v1",
				Kind:       "Deployment",
				Name:       opts.Name,
			},
			MinReplicas: &opts.MinReplicas,
			MaxReplicas: opts.MaxReplicas,
			Metrics: []autoscalingv2.MetricSpec{
				{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricSource{
						Name: "cpu",
						Target: autoscalingv2.MetricTarget{
							Type:               autoscalingv2.UtilizationMetricType,
							AverageUtilization: &opts.CPURate,
						},
					},
				},
			},
		},
	}

	if _, err := clientset.AutoscalingV2().HorizontalPodAutoscalers(opts.Namespace).Create(ctx, hpa, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create hpa resource: %v", err)
		}
		logHandler("hpa resource successfully updated")
	} else {
		logHandler("hpa resource successfully created")
	}

	return nil
}

func DeleteHPA(clientset *kubernetes.Clientset, ctx context.Context, opts HPAOptions, logHandler func(msg string)) error {
	err := clientset.AutoscalingV2().HorizontalPodAutoscalers(opts.Namespace).Delete(ctx, opts.Name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete hpa resource: %v", err)
	}
	if apierrors.IsNotFound(err) {
		logHandler(fmt.Sprintf("hpa resource %s in namespace %s not found, no action taken\n", opts.Name, opts.Namespace))
	} else {
		logHandler(fmt.Sprintf("hpa resource %s in namespace %s successfully deleted\n", opts.Name, opts.Namespace))
	}
	return nil
}
