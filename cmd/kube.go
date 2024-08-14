package cmd

import (
	"context"
	"fmt"

	"github.com/guobinqiu/appdeployer/docker"
	"github.com/guobinqiu/appdeployer/helpers"
	"github.com/guobinqiu/appdeployer/kube"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeOptions struct {
	Kubeconfig        string                 `form:"kubeconfig" json:"kubeconfig"`
	Namespace         string                 `form:"namespace" json:"namespace"`
	IngressOptions    kube.IngressOptions    `form:"ingress" json:"ingress"`
	ServiceOptions    kube.ServiceOptions    `form:"service" json:"service"`
	DeploymentOptions kube.DeploymentOptions `form:"deployment" json:"deployment"`
	HpaOptions        kube.HPAOptions        `form:"hpa" json:"hpa"`
	PvcOptions        kube.PVCOptions        `form:"pvc" json:"pvc"`
}

var dockerOptions docker.DockerOptions
var kubeOptions KubeOptions

func init() {
	// set default values
	viper.SetDefault("docker.dockerconfig", "~/.docker/config.json")
	viper.SetDefault("docker.dockerfile", "./Dockerfile")
	viper.SetDefault("docker.registry", docker.DOCKERHUB)
	viper.SetDefault("docker.tag", "latest")
	viper.SetDefault("kube.kubeconfig", "~/.kube/config")
	viper.SetDefault("kube.ingress.tls", false)
	viper.SetDefault("kube.ingress.selfsigned", false)
	viper.SetDefault("kube.ingress.selfsignedyears", 1)
	viper.SetDefault("kube.service.port", 8000)
	viper.SetDefault("kube.deployment.replicas", 1)
	viper.SetDefault("kube.deployment.port", 8000)
	viper.SetDefault("kube.deployment.rollingupdate.maxsurge", "1")
	viper.SetDefault("kube.deployment.rollingupdate.maxunavailable", "0")
	viper.SetDefault("kube.deployment.livenessprobe.enabled", false)
	viper.SetDefault("kube.deployment.livenessprobe.type", kube.ProbeTypeHTTPGet)
	viper.SetDefault("kube.deployment.livenessprobe.path", "/")
	viper.SetDefault("kube.deployment.livenessprobe.scheme", "http")
	viper.SetDefault("kube.deployment.livenessprobe.initialdelayseconds", 0)
	viper.SetDefault("kube.deployment.livenessprobe.timeoutseconds", 1)
	viper.SetDefault("kube.deployment.livenessprobe.periodseconds", 10)
	viper.SetDefault("kube.deployment.livenessprobe.successthreshold", 1)
	viper.SetDefault("kube.deployment.livenessprobe.failurethreshold", 3)
	viper.SetDefault("kube.deployment.readinessprobe.enabled", false)
	viper.SetDefault("kube.deployment.readinessprobe.type", kube.ProbeTypeHTTPGet)
	viper.SetDefault("kube.deployment.readinessprobe.path", "/")
	viper.SetDefault("kube.deployment.readinessprobe.scheme", "http")
	viper.SetDefault("kube.deployment.readinessprobe.initialdelayseconds", 0)
	viper.SetDefault("kube.deployment.readinessprobe.timeoutseconds", 1)
	viper.SetDefault("kube.deployment.readinessprobe.periodseconds", 10)
	viper.SetDefault("kube.deployment.readinessprobe.successthreshold", 1)
	viper.SetDefault("kube.deployment.readinessprobe.failurethreshold", 3)
	viper.SetDefault("kube.deployment.volumemount.enabled", false)
	viper.SetDefault("kube.deployment.volumemount.mountpath", "/app/data")
	viper.SetDefault("kube.hpa.enabled", false)
	viper.SetDefault("kube.hpa.minreplicas", 1)
	viper.SetDefault("kube.hpa.maxreplicas", 10)
	viper.SetDefault("kube.hpa.cpurate", 50)
	viper.SetDefault("kube.pvc.accessmode", "readwriteonce")
	viper.SetDefault("kube.pvc.storageclassname", "openebs-hostpath")
	viper.SetDefault("kube.pvc.storagesize", "1G")

	// docker
	kubeCmd.Flags().StringVar(&dockerOptions.Dockerconfig, "docker.dockerconfig", viper.GetString("docker.dockerconfig"), "Path to docker configuration. Defaults to ~/.docker/config.json")
	kubeCmd.Flags().StringVar(&dockerOptions.Dockerfile, "docker.dockerfile", viper.GetString("docker.dockerfile"), "Path to Dockerfile for building image. Defaults to appdir/Dockerfile")
	kubeCmd.Flags().StringVar(&dockerOptions.Registry, "docker.registry", viper.GetString("docker.registry"), "URL for docker registry. Defaults to https://index.docker.io/v1/")
	kubeCmd.Flags().StringVar(&dockerOptions.Username, "docker.username", viper.GetString("docker.username"), "Username for docker registry")
	kubeCmd.Flags().StringVar(&dockerOptions.Password, "docker.password", viper.GetString("docker.password"), "Password for docker registry")
	kubeCmd.Flags().StringVar(&dockerOptions.Repository, "docker.repository", viper.GetString("docker.repository"), "Repository for docker registry")
	kubeCmd.Flags().StringVar(&dockerOptions.Tag, "docker.tag", viper.GetString("docker.tag"), "Tag for docker registry. Defaults to latest")

	//kube
	kubeCmd.Flags().StringVar(&kubeOptions.Kubeconfig, "kube.kubeconfig", viper.GetString("kube.kubeconfig"), "Path to kubernetes configuration. Defaults to ~/.kube/config")
	kubeCmd.Flags().StringVar(&kubeOptions.Namespace, "kube.namespace", viper.GetString("kube.namespace"), "Namespace for app resources. Defaults to appname")
	kubeCmd.Flags().StringVar(&kubeOptions.IngressOptions.Host, "kube.ingress.host", viper.GetString("kube.ingress.host"), "Host for app ingress. Defaults to appName.com")
	kubeCmd.Flags().BoolVar(&kubeOptions.IngressOptions.TLS, "kube.ingress.tls", viper.GetBool("kube.ingress.tls"), "Enable or disable TLS for app host. Defaults to false")
	kubeCmd.Flags().BoolVar(&kubeOptions.IngressOptions.SelfSigned, "kube.ingress.selfsigned", viper.GetBool("kube.ingress.selfsigned"), "Enable or disable self-signed certificate. Defaults to false")
	kubeCmd.Flags().IntVar(&kubeOptions.IngressOptions.SelfSignedYears, "kube.ingress.selfsignedyears", viper.GetInt("kube.ingress.selfsignedyears"), "Validity of self-signed certificate. Defaults to 1 year")
	kubeCmd.Flags().StringVar(&kubeOptions.IngressOptions.CrtPath, "kube.ingress.crtpath", viper.GetString("kube.ingress.crtpath"), "Path to .crt file (PEM format) for non self-signed certificate")
	kubeCmd.Flags().StringVar(&kubeOptions.IngressOptions.KeyPath, "kube.ingress.keypath", viper.GetString("kube.ingress.keypath"), "Path to .key file (PEM format) for non self-signed certificate")
	kubeCmd.Flags().Int32Var(&kubeOptions.ServiceOptions.Port, "kube.service.port", viper.GetInt32("kube.service.port"), "Port for app service. Defaults to 8000")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.Replicas, "kube.deployment.replicas", viper.GetInt32("kube.deployment.replicas"), "Number of app pods. Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.Port, "kube.deployment.port", viper.GetInt32("kube.deployment.port"), "Container port for each app pod. Defaults to 8000, as same as service port")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.RollingUpdate.MaxSurge, "kube.deployment.rollingupdate.maxsurge", viper.GetString("kube.deployment.rollingupdate.maxsurge"), "MaxSurge for rolling update app pods. Defaults to 1")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.RollingUpdate.MaxUnavailable, "kube.deployment.rollingupdate.maxunavailable", viper.GetString("kube.deployment.rollingupdate.maxunavailable"), "MaxUnavailable for rolling update app pods. Defaults to 0")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.Quota.CPULimit, "kube.deployment.quota.cpulimit", viper.GetString("kube.deployment.quota.cpulimit"), "CPU limit for each app container (one pod one container)")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.Quota.MemLimit, "kube.deployment.quota.memlimit", viper.GetString("kube.deployment.quota.memlimit"), "Memory limit for each app container (one pod one container)")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.Quota.CPURequest, "kube.deployment.quota.cpurequest", viper.GetString("kube.deployment.quota.cpurequest"), "CPU request for each app container (one pod one container)")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.Quota.MemRequest, "kube.deployment.quota.memrequest", viper.GetString("kube.deployment.quota.memrequest"), "Memory request for each app container (one pod one container)")
	kubeCmd.Flags().BoolVar(&kubeOptions.DeploymentOptions.LivenessProbe.Enabled, "kube.deployment.livenessprobe.enabled", viper.GetBool("kube.deployment.livenessprobe.enabled"), "Enable or disable liveness probe for each app container (one pod one container). Defaults to false")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.LivenessProbe.Type, "kube.deployment.livenessprobe.type", viper.GetString("kube.deployment.livenessprobe.type"), "Type of liveness probe for each app container (one pod one container). Such as HTTPGet, TCPSocket and Exec. Defaults to HTTPGet")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.LivenessProbe.Path, "kube.deployment.livenessprobe.path", viper.GetString("kube.deployment.livenessprobe.path"), "Path of liveness probe for each app container (one pod one container). Correspond to HTTPGet type. Defaults to /")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.LivenessProbe.Scheme, "kube.deployment.livenessprobe.scheme", viper.GetString("kube.deployment.livenessprobe.scheme"), "Scheme of liveness probe for each app container (one pod one container). Correspond to HTTPGet type. Such as HTTP and HTTPS. Defaults to HTTP")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.LivenessProbe.Command, "kube.deployment.livenessprobe.command", viper.GetString("kube.deployment.livenessprobe.command"), "Command of liveness probe for each app container (one pod one container). Correspond to Exec type")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.LivenessProbe.InitialDelaySeconds, "kube.deployment.livenessprobe.initialdelayseconds", viper.GetInt32("kube.deployment.livenessprobe.initialdelayseconds"), "Initial delay seconds of liveness probe for each app container (one pod one container). Defaults to 0")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.LivenessProbe.TimeoutSeconds, "kube.deployment.livenessprobe.timeoutseconds", viper.GetInt32("kube.deployment.livenessprobe.timeoutseconds"), "Timeout seconds of liveness probe for each app container (one pod one container). Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.LivenessProbe.PeriodSeconds, "kube.deployment.livenessprobe.periodseconds", viper.GetInt32("kube.deployment.livenessprobe.periodseconds"), "Period seconds of liveness probe for each app container (one pod one container). Defaults to 10")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.LivenessProbe.SuccessThreshold, "kube.deployment.livenessprobe.successthreshold", viper.GetInt32("kube.deployment.livenessprobe.successthreshold"), "Success threshold of liveness probe for each app container (one pod one container). Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.LivenessProbe.FailureThreshold, "kube.deployment.livenessprobe.failurethreshold", viper.GetInt32("kube.deployment.livenessprobe.failurethreshold"), "Failure threshold of liveness probe for each app container (one pod one container). Defaults to 3")
	kubeCmd.Flags().BoolVar(&kubeOptions.DeploymentOptions.ReadinessProbe.Enabled, "kube.deployment.readinessprobe.enabled", viper.GetBool("kube.deployment.readinessprobe.enabled"), "Enable or disable readiness probe for each app container (one pod one container)")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.ReadinessProbe.Type, "kube.deployment.readinessprobe.type", viper.GetString("kube.deployment.readinessprobe.type"), "Type of readiness probe for each app container (one pod one container). Such as HTTPGet, TCPSocket and Exec. Defaults to HTTPGet")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.ReadinessProbe.Path, "kube.deployment.readinessprobe.path", viper.GetString("kube.deployment.readinessprobe.path"), "Path of readiness probe for each app container (one pod one container). Correspond to HTTPGet type. Defaults to /")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.ReadinessProbe.Scheme, "kube.deployment.readinessprobe.scheme", viper.GetString("kube.deployment.readinessprobe.scheme"), "Scheme of readiness probe for each app container (one pod one container). Correspond to HTTPGet type. Such as HTTP and HTTPS. Defaults to HTTP")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.ReadinessProbe.Command, "kube.deployment.readinessprobe.command", viper.GetString("kube.deployment.readinessprobe.command"), "Command of readiness probe for each app container (one pod one container). Correspond to Exec type")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.ReadinessProbe.InitialDelaySeconds, "kube.deployment.readinessprobe.initialdelayseconds", viper.GetInt32("kube.deployment.readinessprobe.initialdelayseconds"), "Initial delay seconds of readiness probe for each app container (one pod one container). Defaults to 0")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.ReadinessProbe.TimeoutSeconds, "kube.deployment.readinessprobe.timeoutseconds", viper.GetInt32("kube.deployment.readinessprobe.timeoutseconds"), "Timeout seconds of readiness probe for each app container (one pod one container). Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.ReadinessProbe.PeriodSeconds, "kube.deployment.readinessprobe.periodseconds", viper.GetInt32("kube.deployment.readinessprobe.periodseconds"), "Period seconds of readiness probe for each app container (one pod one container). Defaults to 10")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.ReadinessProbe.SuccessThreshold, "kube.deployment.readinessprobe.successthreshold", viper.GetInt32("kube.deployment.readinessprobe.successthreshold"), "Success threshold of readiness probe for each app container (one pod one container). Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.DeploymentOptions.ReadinessProbe.FailureThreshold, "kube.deployment.readinessprobe.failurethreshold", viper.GetInt32("kube.deployment.readinessprobe.failurethreshold"), "Failure threshold of readiness probe for each app container (one pod one container). Defaults to 3")
	kubeCmd.Flags().BoolVar(&kubeOptions.DeploymentOptions.VolumeMount.Enabled, "kube.deployment.volumemount.enabled", viper.GetBool("kube.deployment.volumemount.enabled"), "Enable or disable volume mount for each app pod. Defaults to false")
	kubeCmd.Flags().StringVar(&kubeOptions.DeploymentOptions.VolumeMount.MountPath, "kube.deployment.volumemount.mountpath", viper.GetString("kube.deployment.volumemount.mountpath"), "Path of volume mount for each app pod. Defaults to /app/data")
	kubeCmd.Flags().BoolVar(&kubeOptions.HpaOptions.Enabled, "kube.hpa.enabled", viper.GetBool("kube.hpa.enabled"), "Enable or disable HPA (Horizontal Pod Autoscaler) for app pods. Defaults to false")
	kubeCmd.Flags().Int32Var(&kubeOptions.HpaOptions.MinReplicas, "kube.hpa.minreplicas", viper.GetInt32("kube.hpa.minreplicas"), "Number of minimum pods for HPA (Horizontal Pod Autoscaler). Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.HpaOptions.MaxReplicas, "kube.hpa.maxreplicas", viper.GetInt32("kube.hpa.maxreplicas"), "Number of maximum pods for HPA (Horizontal Pod Autoscaler). Defaults to 10")
	kubeCmd.Flags().Int32Var(&kubeOptions.HpaOptions.CPURate, "kube.hpa.cpurate", viper.GetInt32("kube.hpa.cpurate"), "Average CPU utilization for HPA (Horizontal Pod Autoscaler). Defaults to 50")
	kubeCmd.Flags().StringVar(&kubeOptions.PvcOptions.AccessMode, "kube.pvc.accessmode", viper.GetString("kube.pvc.accessmode"), "Access mode of persistent storage for pod volumn mount. Such as ReadWriteOnce, ReadOnlyMany and ReadWriteMany. Defaults to ReadWriteOnce")
	kubeCmd.Flags().StringVar(&kubeOptions.PvcOptions.StorageClassName, "kube.pvc.storageclassname", viper.GetString("kube.pvc.storageclassname"), "Classname of persistent storage for pod volumn mount. Defaults to openebs-hostpath")
	kubeCmd.Flags().StringVar(&kubeOptions.PvcOptions.StorageSize, "kube.pvc.storagesize", viper.GetString("kube.pvc.storagesize"), "Size of persistent storage for pod volumn mount. Defaults to 1G")
	kubeCmd.Flags().StringSliceVarP(&kubeOptions.DeploymentOptions.EnvVars, "env", "e", nil, "Set environment variables in the form of key=value")
}

var kubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "Deploy app to kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		setDefaultOptions()
		setDockerOptions()
		setKubeOptions()

		gitPull()

		// Create a docker service
		dockerservice, err := docker.NewDockerService()
		if err != nil {
			panic(err)
		}

		//TODO handle timeout or cancel
		ctx := context.TODO()

		// Build an app into a docker image
		if err := dockerservice.BuildImage(ctx, dockerOptions); err != nil {
			panic(err)
		}

		// Push the docker image to docker registry
		if err := dockerservice.PushImage(ctx, dockerOptions); err != nil {
			panic(err)
		}

		dockerservice.Close()

		// Create a kubernetes client by the specified kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", kubeOptions.Kubeconfig)
		if err != nil {
			panic(err)
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}

		// Update or create kubernetes resource objects
		if err := kube.CreateOrUpdateNamespace(clientset, ctx, kubeOptions.Namespace); err != nil {
			panic(err)
		}

		if err := kube.CreateOrUpdateDockerSecret(clientset, ctx, kube.DockerSecretOptions{
			Name:          defaultOptions.AppName,
			Namespace:     kubeOptions.Namespace,
			DockerOptions: dockerOptions,
		}); err != nil {
			panic(err)
		}

		if err := kube.CreateOrUpdateServiceAccount(clientset, ctx, kube.ServiceAccountOptions{
			Name:      defaultOptions.AppName,
			Namespace: kubeOptions.Namespace,
		}); err != nil {
			panic(err)
		}

		if kubeOptions.DeploymentOptions.VolumeMount.Enabled {
			kubeOptions.PvcOptions.Name = defaultOptions.AppName
			kubeOptions.PvcOptions.Namespace = kubeOptions.Namespace
			if err := kube.CreateOrUpdatePVC(clientset, ctx, kubeOptions.PvcOptions); err != nil {
				panic(err)
			}
		} else {
			kubeOptions.DeploymentOptions.Name = defaultOptions.AppName
			kubeOptions.DeploymentOptions.Namespace = kubeOptions.Namespace
			kubeOptions.DeploymentOptions.Image = dockerOptions.Image()
			if err := kube.DeleteDeployment(clientset, ctx, kubeOptions.DeploymentOptions); err != nil {
				panic(err)
			}

			kubeOptions.HpaOptions.Name = defaultOptions.AppName
			kubeOptions.HpaOptions.Namespace = kubeOptions.Namespace
			if err := kube.DeletePVC(clientset, ctx, kubeOptions.HpaOptions); err != nil {
				panic(err)
			}
		}

		kubeOptions.DeploymentOptions.Name = defaultOptions.AppName
		kubeOptions.DeploymentOptions.Namespace = kubeOptions.Namespace
		kubeOptions.DeploymentOptions.Image = dockerOptions.Image()
		if err := kube.CreateOrUpdateDeployment(clientset, ctx, kubeOptions.DeploymentOptions); err != nil {
			panic(err)
		}

		kubeOptions.ServiceOptions.Name = defaultOptions.AppName
		kubeOptions.ServiceOptions.Namespace = kubeOptions.Namespace
		kubeOptions.ServiceOptions.TargetPort = kubeOptions.DeploymentOptions.Port
		if err := kube.CreateOrUpdateService(clientset, ctx, kubeOptions.ServiceOptions); err != nil {
			panic(err)
		}

		kubeOptions.IngressOptions.Name = defaultOptions.AppName
		kubeOptions.IngressOptions.Namespace = kubeOptions.Namespace
		if err := kube.CreateOrUpdateIngress(clientset, ctx, kubeOptions.IngressOptions); err != nil {
			panic(err)
		}

		if kubeOptions.HpaOptions.Enabled {
			kubeOptions.HpaOptions.Name = defaultOptions.AppName
			kubeOptions.HpaOptions.Namespace = kubeOptions.Namespace
			if err := kube.CreateOrUpdateHPA(clientset, ctx, kubeOptions.HpaOptions); err != nil {
				panic(err)
			}
		} else {
			kubeOptions.HpaOptions.Name = defaultOptions.AppName
			kubeOptions.HpaOptions.Namespace = kubeOptions.Namespace
			if err := kube.DeleteHPA(clientset, ctx, kubeOptions.HpaOptions); err != nil {
				panic(err)
			}
		}
	},
}

func setDockerOptions() {
	dockerOptions.AppDir = defaultOptions.AppDir

	dockerOptions.Dockerconfig = helpers.ExpandUser(dockerOptions.Dockerconfig)
	exist, err := helpers.IsFileExist(dockerOptions.Dockerconfig)
	if err != nil {
		panic(err)
	}
	if !exist {
		panic("dockerconfig does not exist")
	}

	if helpers.IsBlank(dockerOptions.Repository) && dockerOptions.Registry == docker.DOCKERHUB {
		if helpers.IsBlank(dockerOptions.Username) {
			panic("docker.username is required")
		}
		dockerOptions.Repository = fmt.Sprintf("%s/%s", dockerOptions.Username, defaultOptions.AppName)
	}
}

func setKubeOptions() {
	kubeOptions.Kubeconfig = helpers.ExpandUser(kubeOptions.Kubeconfig)
	exist, err := helpers.IsFileExist(kubeOptions.Kubeconfig)
	if err != nil {
		panic(err)
	}
	if !exist {
		panic("kubeconfig does not exist")
	}

	if helpers.IsBlank(kubeOptions.Namespace) {
		kubeOptions.Namespace = defaultOptions.AppName
	}

	if helpers.IsBlank(kubeOptions.IngressOptions.Host) {
		kubeOptions.IngressOptions.Host = fmt.Sprintf("%s.com", defaultOptions.AppName)
	}

	if kubeOptions.IngressOptions.TLS && !kubeOptions.IngressOptions.SelfSigned {
		if helpers.IsBlank(kubeOptions.IngressOptions.CrtPath) {
			panic("crt path does not exist")
		}
		if helpers.IsBlank(kubeOptions.IngressOptions.KeyPath) {
			panic("key path does not exist")
		}
	}
}
