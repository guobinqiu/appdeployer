package kube

import (
	"context"
	"fmt"

	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type HPAOptions struct {
	Name        string
	Namespace   string
	Enabled     bool
	MinReplicas int32
	MaxReplicas int32
	CPURate     int32
}

func CreateOrUpdateHPA(clientset *kubernetes.Clientset, ctx context.Context, opts HPAOptions) error {
	hpa := &autoscalingv2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.Namespace,
		},
		Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2beta2.CrossVersionObjectReference{
				APIVersion: "app/v1",
				Kind:       "Deployment",
				Name:       opts.Name,
			},
			MinReplicas: &opts.MinReplicas,
			MaxReplicas: opts.MaxReplicas,
			Metrics: []autoscalingv2beta2.MetricSpec{
				{
					Type: autoscalingv2beta2.ResourceMetricSourceType,
					Resource: &autoscalingv2beta2.ResourceMetricSource{
						Name: "cpu",
						Target: autoscalingv2beta2.MetricTarget{
							Type:               autoscalingv2beta2.UtilizationMetricType,
							AverageUtilization: &opts.CPURate,
						},
					},
				},
			},
		},
	}

	if _, err := clientset.AutoscalingV2beta2().HorizontalPodAutoscalers(opts.Namespace).Create(ctx, hpa, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create hpa resource: %v", err)
		}
		fmt.Println("hpa resource successfully updated")
	} else {
		fmt.Println("hpa resource successfully created")
	}

	return nil
}

func DeleteHPA(clientset *kubernetes.Clientset, ctx context.Context, opts HPAOptions) error {
	err := clientset.AutoscalingV2beta2().HorizontalPodAutoscalers(opts.Namespace).Delete(ctx, opts.Name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete hpa resource: %v", err)
	}
	if apierrors.IsNotFound(err) {
		fmt.Printf("hpa %s in namespace %s not found, no action taken\n", opts.Name, opts.Namespace)
	} else {
		fmt.Printf("hpa resource %s in namespace %s successfully deleted\n", opts.Name, opts.Namespace)
	}
	return nil
}
