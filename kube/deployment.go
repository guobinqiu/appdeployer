package kube

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/guobinqiu/appdeployer/helpers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// DeploymentOptions 用于配置 Deployment 创建或更新的选项
type DeploymentOptions struct {
	Name           string
	Namespace      string
	Replicas       int32
	Image          string
	Port           int32
	MaxSurge       string
	MaxUnavailable string
	CPURequest     string
	CPULimit       string
	MemRequest     string
	MemLimit       string
}

func CreateOrUpdateDeployment(clientset *kubernetes.Clientset, ctx context.Context, opts DeploymentOptions) error {
	maxSurge := intstr.Parse(opts.MaxSurge)
	maxUnavailable := intstr.Parse(opts.MaxUnavailable)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.Namespace,
		},

		Spec: appsv1.DeploymentSpec{
			Replicas: &opts.Replicas,

			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &maxSurge,
					MaxUnavailable: &maxUnavailable,
				},
			},

			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": opts.Name,
				},
			},

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": opts.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            opts.Name,
							Image:           opts.Image,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: opts.Port,
								},
							},
						},
					},
					ServiceAccountName: opts.Name,
				},
			},
		},
	}

	limits := corev1.ResourceList{}
	if !helpers.IsBlank(opts.CPULimit) {
		cpuLimit, err := parseCPUSize(strings.ToLower(opts.CPULimit))
		if err != nil {
			return err
		}
		limits[corev1.ResourceCPU] = *cpuLimit
	}
	if !helpers.IsBlank(opts.MemLimit) {
		memLimit, err := parseMemorySize(strings.ToLower(opts.MemLimit))
		if err != nil {
			return err
		}
		limits[corev1.ResourceMemory] = *memLimit
	}

	requests := corev1.ResourceList{}
	if !helpers.IsBlank(opts.CPURequest) {
		cpuRequest, err := parseCPUSize(strings.ToLower(opts.CPURequest))
		if err != nil {
			return err
		}
		requests[corev1.ResourceCPU] = *cpuRequest
	}
	if !helpers.IsBlank(opts.MemRequest) {
		memRequest, err := parseMemorySize(strings.ToLower(opts.MemRequest))
		if err != nil {
			return err
		}
		requests[corev1.ResourceMemory] = *memRequest
	}

	if len(limits) > 0 || len(requests) > 0 {
		resource := corev1.ResourceRequirements{}
		if len(limits) > 0 {
			resource.Limits = limits
		}
		if len(requests) > 0 {
			resource.Requests = requests
		}
		deployment.Spec.Template.Spec.Containers[0].Resources = resource
	}

	_, err := clientset.AppsV1().Deployments(opts.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create deployment resource: %v", err)
		}
		fmt.Println("deployment resource already exists, attempting update...")
		_, err := clientset.AppsV1().Deployments(opts.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update deployment resource: %v", err)
		}
		fmt.Println("deployment resource successfully updated")
	} else {
		fmt.Println("deployment resource successfully created")
	}

	return nil
}

func parseCPUSize(input string) (*resource.Quantity, error) {
	re, err := regexp.Compile(`^(\d+)m$`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return nil, fmt.Errorf("invalid cpu format, expected 'number{m}', got '%s'", input)
	}

	value, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpu value: %w", err)
	}

	return resource.NewMilliQuantity(value, resource.DecimalSI), nil
}

func parseMemorySize(input string) (*resource.Quantity, error) {
	re, err := regexp.Compile(`^(\d+)([kmg]i)$`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return nil, fmt.Errorf("invalid memory format, expected 'number{Ki|Mi|Gi}', got '%s'", input)
	}

	value, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse memory value: %w", err)
	}

	unit := matches[2]
	switch unit {
	case "ki":
		return resource.NewQuantity(value*(1<<10), resource.BinarySI), nil
	case "mi":
		return resource.NewQuantity(value*(1<<20), resource.BinarySI), nil
	case "gi":
		return resource.NewQuantity(value*(1<<30), resource.BinarySI), nil
	default:
		return nil, fmt.Errorf("invalid memory unit '%s'", unit)
	}
}
