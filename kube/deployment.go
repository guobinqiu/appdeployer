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

const (
	ProbeTypeHTTPGet   = "httpget"
	ProbeTypeExec      = "exec"
	ProbeTypeTCPSocket = "tcpsocket"
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
	EnvVars        []string
	LivenessProbe  LivenessProbe
	ReadinessProbe ReadinessProbe
}

type LivenessProbe struct {
	Enabled bool
	Type    string
	Path    string
	Port    string
	Schema  string
	Command string
	ProbeParams
}

type ReadinessProbe struct {
	Enabled bool
	Type    string
	Path    string
	Port    string
	Schema  string
	Command string
	ProbeParams
}

type ProbeParams struct {
	InitialDelaySeconds int32
	TimeoutSeconds      int32
	PeriodSeconds       int32
	SuccessThreshold    int32
	FailureThreshold    int32
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

	container := deployment.Spec.Template.Spec.Containers[0]
	if err := setResource(&container, opts); err != nil {
		return fmt.Errorf("failed to set resource: %v", err)
	}
	if err := setLivenessProbe(&container, opts); err != nil {
		return fmt.Errorf("failed to set liveness probe: %v", err)
	}
	if err := setReadinessProbe(&container, opts); err != nil {
		return fmt.Errorf("failed to set readiness probe: %v", err)
	}
	if err := setEnv(&container, opts); err != nil {
		return fmt.Errorf("failed to set env: %v", err)
	}
	deployment.Spec.Template.Spec.Containers[0] = container

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

	var bytesValue int64
	unit := matches[2]
	switch unit {
	case "ki":
		bytesValue = value << 10 //1KiB = 2^10B
	case "mi":
		bytesValue = value << 20 //1MiB = 2^20B
	case "gi":
		bytesValue = value << 30 //1GiB = 2^30B
	default:
		return nil, fmt.Errorf("unsupported memory unit: '%s'", unit)
	}
	return resource.NewQuantity(bytesValue, resource.BinarySI), nil
}

func setResource(container *corev1.Container, opts DeploymentOptions) error {
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
		container.Resources = resource
	}

	return nil
}

func setLivenessProbe(container *corev1.Container, opts DeploymentOptions) error {
	if !opts.LivenessProbe.Enabled {
		return nil
	}

	var probe Probe
	probeType := strings.ToLower(opts.LivenessProbe.Type)
	switch probeType {
	case ProbeTypeHTTPGet:
		probe = HttpGetProbe{
			Path:   opts.LivenessProbe.Path,
			Port:   intstr.FromInt32(opts.Port),
			Schema: corev1.URIScheme(strings.ToUpper(opts.LivenessProbe.Schema)),
		}
	case ProbeTypeExec:
		probe = ExecProbe{
			Command: opts.LivenessProbe.Command,
		}
	case ProbeTypeTCPSocket:
		probe = TCPSocketProbe{
			Port: intstr.FromInt32(opts.Port),
		}
	default:
		return fmt.Errorf("unsupported liveness probe type: '%s'", probeType)
	}

	container.LivenessProbe = probe.GetProbe()
	return nil
}

func setReadinessProbe(container *corev1.Container, opts DeploymentOptions) error {
	if !opts.ReadinessProbe.Enabled {
		return nil
	}

	var probe Probe
	probeType := strings.ToLower(opts.ReadinessProbe.Type)
	switch probeType {
	case ProbeTypeHTTPGet:
		probe = HttpGetProbe{
			Path:   opts.ReadinessProbe.Path,
			Port:   intstr.FromInt32(opts.Port),
			Schema: corev1.URIScheme(strings.ToUpper(opts.ReadinessProbe.Schema)),
		}
	case ProbeTypeExec:
		probe = ExecProbe{
			Command: opts.ReadinessProbe.Command,
		}
	case ProbeTypeTCPSocket:
		probe = TCPSocketProbe{
			Port: intstr.FromInt32(opts.Port),
		}
	default:
		return fmt.Errorf("unsupported readiness probe type: '%s'", probeType)
	}

	container.ReadinessProbe = probe.GetProbe()
	return nil
}

func setEnv(container *corev1.Container, opts DeploymentOptions) error {
	var envs []corev1.EnvVar
	for _, envVar := range opts.EnvVars {
		parts := strings.Split(envVar, "=")
		if len(parts) != 2 {
			return fmt.Errorf("invalid format for environment variable: '%s'", envVar)
		}
		envs = append(envs, corev1.EnvVar{
			Name:  parts[0],
			Value: parts[1],
		})
	}
	if len(envs) > 0 {
		container.Env = envs
	}
	return nil
}
