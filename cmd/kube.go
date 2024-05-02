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
	Kubeconfig        string
	Namespace         string
	ingressOptions    kube.IngressOptions
	serviceOptions    kube.ServiceOptions
	deploymentOptions kube.DeploymentOptions
	hpaOptions        kube.HPAOptions
	pvcOptions        kube.PVCOptions
}

var dockerOptions docker.DockerOptions
var kubeOptions KubeOptions

func init() {
	// set default values
	viper.SetDefault("docker.dockerconfig", helpers.GetDefaultDockerConfig())
	viper.SetDefault("docker.dockerfile", "./Dockerfile")
	viper.SetDefault("docker.registry", docker.DOCKERHUB)
	viper.SetDefault("docker.tag", "latest")
	viper.SetDefault("kube.kubeconfig", helpers.GetDefaultKubeConfig())
	viper.SetDefault("kube.ingress.tls", true)
	viper.SetDefault("kube.service.port", 8000)
	viper.SetDefault("kube.deployment.replicas", 1)
	viper.SetDefault("kube.deployment.port", 8000)
	viper.SetDefault("kube.deployment.maxsurge", "1")
	viper.SetDefault("kube.deployment.maxunavailable", "0")

	viper.SetDefault("kube.deployment.livenessprobe.enabled", false)
	viper.SetDefault("kube.deployment.livenessprobe.type", kube.ProbeTypeHTTPGet)
	viper.SetDefault("kube.deployment.livenessprobe.path", "/")
	viper.SetDefault("kube.deployment.livenessprobe.schema", "http")
	viper.SetDefault("kube.deployment.livenessprobe.initialdelayseconds", 0)
	viper.SetDefault("kube.deployment.livenessprobe.timeoutseconds", 1)
	viper.SetDefault("kube.deployment.livenessprobe.periodseconds", 10)
	viper.SetDefault("kube.deployment.livenessprobe.successthreshold", 1)
	viper.SetDefault("kube.deployment.livenessprobe.failurethreshold", 3)

	viper.SetDefault("kube.deployment.readinessprobe.enabled", false)
	viper.SetDefault("kube.deployment.readinessprobe.type", kube.ProbeTypeHTTPGet)
	viper.SetDefault("kube.deployment.readinessprobe.path", "/")
	viper.SetDefault("kube.deployment.readinessprobe.schema", "http")
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
	kubeCmd.Flags().StringVar(&dockerOptions.Dockerconfig, "docker.dockerconfig", viper.GetString("docker.dockerconfig"), "docker.dockerconfig")
	kubeCmd.Flags().StringVar(&dockerOptions.Dockerfile, "docker.dockerfile", viper.GetString("docker.dockerfile"), "docker.dockerfile")
	kubeCmd.Flags().StringVar(&dockerOptions.Registry, "docker.registry", viper.GetString("docker.registry"), "docker.registry")
	kubeCmd.Flags().StringVar(&dockerOptions.Username, "docker.username", viper.GetString("docker.username"), "docker.username")
	kubeCmd.Flags().StringVar(&dockerOptions.Password, "docker.password", viper.GetString("docker.password"), "docker.password")
	kubeCmd.Flags().StringVar(&dockerOptions.Repository, "docker.repository", viper.GetString("docker.repository"), "docker.repository")
	kubeCmd.Flags().StringVar(&dockerOptions.Tag, "docker.tag", viper.GetString("docker.tag"), "docker.tag")

	//kube
	kubeCmd.Flags().StringVar(&kubeOptions.Kubeconfig, "kube.kubeconfig", viper.GetString("kube.kubeconfig"), "kube.kubeconfig")
	kubeCmd.Flags().StringVar(&kubeOptions.Namespace, "kube.namespace", viper.GetString("kube.namespace"), "kube.namespace")
	kubeCmd.Flags().StringVar(&kubeOptions.ingressOptions.Host, "kube.ingress.host", viper.GetString("kube.ingress.host"), "kube.ingress.host")
	kubeCmd.Flags().BoolVar(&kubeOptions.ingressOptions.TLS, "kube.ingress.tls", viper.GetBool("kube.ingress.tls"), "kube.ingress.tls")
	kubeCmd.Flags().Int32Var(&kubeOptions.serviceOptions.Port, "kube.service.port", viper.GetInt32("kube.service.port"), "kube.service.port")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.Replicas, "kube.deployment.replicas", viper.GetInt32("kube.deployment.replicas"), "kube.deployment.replicas")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.Port, "kube.deployment.port", viper.GetInt32("kube.deployment.port"), "kube.deployment.port")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.MaxSurge, "kube.deployment.maxsurge", viper.GetString("kube.deployment.maxsurge"), "kube.deployment.maxsurge")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.MaxUnavailable, "kube.deployment.maxunavailable", viper.GetString("kube.deployment.maxunavailable"), "kube.deployment.maxunavailable")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.CPULimit, "kube.deployment.cpulimit", viper.GetString("kube.deployment.cpulimit"), "kube.deployment.cpulimit")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.MemLimit, "kube.deployment.memlimit", viper.GetString("kube.deployment.memlimit"), "kube.deployment.memlimit")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.CPURequest, "kube.deployment.cpurequest", viper.GetString("kube.deployment.cpurequest"), "kube.deployment.cpurequest")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.MemRequest, "kube.deployment.memrequest", viper.GetString("kube.deployment.memrequest"), "kube.deployment.memrequest")

	kubeCmd.Flags().StringSliceVarP(&kubeOptions.deploymentOptions.EnvVars, "env", "e", nil, "Set environment variables in the form of key=value")

	kubeCmd.Flags().BoolVar(&kubeOptions.deploymentOptions.LivenessProbe.Enabled, "kube.deployment.livenessprobe.enabled", viper.GetBool("kube.deployment.livenessprobe.enabled"), "kube.deployment.livenessprobe.enabled")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.LivenessProbe.Type, "kube.deployment.livenessprobe.type", viper.GetString("kube.deployment.livenessprobe.type"), "kube.deployment.livenessprobe.type")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.LivenessProbe.Path, "kube.deployment.livenessprobe.path", viper.GetString("kube.deployment.livenessprobe.path"), "kube.deployment.livenessprobe.path")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.LivenessProbe.Schema, "kube.deployment.livenessprobe.schema", viper.GetString("kube.deployment.livenessprobe.schema"), "kube.deployment.livenessprobe.schema")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.LivenessProbe.Command, "kube.deployment.livenessprobe.command", viper.GetString("kube.deployment.livenessprobe.command"), "kube.deployment.livenessprobe.command")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.LivenessProbe.InitialDelaySeconds, "kube.deployment.livenessprobe.initialdelayseconds", viper.GetInt32("kube.deployment.livenessprobe.initialdelayseconds"), "kube.deployment.livenessprobe.initialdelayseconds")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.LivenessProbe.TimeoutSeconds, "kube.deployment.livenessprobe.timeoutseconds", viper.GetInt32("kube.deployment.livenessprobe.timeoutseconds"), "kube.deployment.livenessprobe.timeoutseconds")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.LivenessProbe.PeriodSeconds, "kube.deployment.livenessprobe.periodseconds", viper.GetInt32("kube.deployment.livenessprobe.periodseconds"), "kube.deployment.livenessprobe.periodseconds")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.LivenessProbe.SuccessThreshold, "kube.deployment.livenessprobe.successthreshold", viper.GetInt32("kube.deployment.livenessprobe.successthreshold"), "kube.deployment.livenessprobe.successthreshold")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.LivenessProbe.FailureThreshold, "kube.deployment.livenessprobe.failurethreshold", viper.GetInt32("kube.deployment.livenessprobe.failurethreshold"), "kube.deployment.livenessprobe.failurethreshold")

	kubeCmd.Flags().BoolVar(&kubeOptions.deploymentOptions.ReadinessProbe.Enabled, "kube.deployment.readinessprobe.enabled", viper.GetBool("kube.deployment.readinessprobe.enabled"), "kube.deployment.readinessprobe.enabled")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.ReadinessProbe.Type, "kube.deployment.readinessprobe.type", viper.GetString("kube.deployment.readinessprobe.type"), "kube.deployment.readinessprobe.type")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.ReadinessProbe.Path, "kube.deployment.readinessprobe.path", viper.GetString("kube.deployment.readinessprobe.path"), "kube.deployment.readinessprobe.path")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.ReadinessProbe.Schema, "kube.deployment.readinessprobe.schema", viper.GetString("kube.deployment.readinessprobe.schema"), "kube.deployment.readinessprobe.schema")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.ReadinessProbe.Command, "kube.deployment.readinessprobe.command", viper.GetString("kube.deployment.readinessprobe.command"), "kube.deployment.readinessprobe.command")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.ReadinessProbe.InitialDelaySeconds, "kube.deployment.readinessprobe.initialdelayseconds", viper.GetInt32("kube.deployment.readinessprobe.initialdelayseconds"), "kube.deployment.readinessprobe.initialdelayseconds")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.ReadinessProbe.TimeoutSeconds, "kube.deployment.readinessprobe.timeoutseconds", viper.GetInt32("kube.deployment.readinessprobe.timeoutseconds"), "kube.deployment.readinessprobe.timeoutseconds")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.ReadinessProbe.PeriodSeconds, "kube.deployment.readinessprobe.periodseconds", viper.GetInt32("kube.deployment.readinessprobe.periodseconds"), "kube.deployment.readinessprobe.periodseconds")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.ReadinessProbe.SuccessThreshold, "kube.deployment.readinessprobe.successthreshold", viper.GetInt32("kube.deployment.readinessprobe.successthreshold"), "kube.deployment.readinessprobe.successthreshold")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.ReadinessProbe.FailureThreshold, "kube.deployment.readinessprobe.failurethreshold", viper.GetInt32("kube.deployment.readinessprobe.failurethreshold"), "kube.deployment.readinessprobe.failurethreshold")

	kubeCmd.Flags().BoolVar(&kubeOptions.deploymentOptions.VolumeMount.Enabled, "kube.deployment.volumemount.enabled", viper.GetBool("kube.deployment.volumemount.enabled"), "kube.deployment.volumemount.enabled")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.VolumeMount.MountPath, "kube.deployment.volumemount.mountpath", viper.GetString("kube.deployment.volumemount.mountpath"), "kube.deployment.volumemount.mountpath")

	kubeCmd.Flags().BoolVar(&kubeOptions.hpaOptions.Enabled, "kube.hpa.enabled", viper.GetBool("kube.hpa.enabled"), "kube.hpa.enabled")
	kubeCmd.Flags().Int32Var(&kubeOptions.hpaOptions.MinReplicas, "kube.hpa.minreplicas", viper.GetInt32("kube.hpa.minreplicas"), "kube.hpa.minreplicas")
	kubeCmd.Flags().Int32Var(&kubeOptions.hpaOptions.MaxReplicas, "kube.hpa.maxreplicas", viper.GetInt32("kube.hpa.maxreplicas"), "kube.hpa.maxreplicas")
	kubeCmd.Flags().Int32Var(&kubeOptions.hpaOptions.CPURate, "kube.hpa.cpurate", viper.GetInt32("kube.hpa.cpurate"), "kube.hpa.cpurate")

	kubeCmd.Flags().StringVar(&kubeOptions.pvcOptions.AccessMode, "kube.pvc.accessmode", viper.GetString("kube.pvc.accessmode"), "kube.pvc.accessmode")
	kubeCmd.Flags().StringVar(&kubeOptions.pvcOptions.StorageClassName, "kube.pvc.storageclassname", viper.GetString("kube.pvc.storageclassname"), "kube.pvc.storageclassname")
	kubeCmd.Flags().StringVar(&kubeOptions.pvcOptions.StorageSize, "kube.pvc.storagesize", viper.GetString("kube.pvc.storagesize"), "kube.pvc.storagesize")
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

		// Update (or create if not exists) kubernetes resource objects
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

		if kubeOptions.deploymentOptions.VolumeMount.Enabled {
			kubeOptions.pvcOptions.Name = defaultOptions.AppName
			kubeOptions.pvcOptions.Namespace = kubeOptions.Namespace
			if err := kube.CreateOrUpdatePVC(clientset, ctx, kubeOptions.pvcOptions); err != nil {
				panic(err)
			}
		} else {
			kubeOptions.deploymentOptions.Name = defaultOptions.AppName
			kubeOptions.deploymentOptions.Namespace = kubeOptions.Namespace
			kubeOptions.deploymentOptions.Image = dockerOptions.Image()
			if err := kube.DeleteDeployment(clientset, ctx, kubeOptions.deploymentOptions); err != nil {
				panic(err)
			}

			kubeOptions.hpaOptions.Name = defaultOptions.AppName
			kubeOptions.hpaOptions.Namespace = kubeOptions.Namespace
			if err := kube.DeletePVC(clientset, ctx, kubeOptions.hpaOptions); err != nil {
				panic(err)
			}
		}

		kubeOptions.deploymentOptions.Name = defaultOptions.AppName
		kubeOptions.deploymentOptions.Namespace = kubeOptions.Namespace
		kubeOptions.deploymentOptions.Image = dockerOptions.Image()
		if err := kube.CreateOrUpdateDeployment(clientset, ctx, kubeOptions.deploymentOptions); err != nil {
			panic(err)
		}

		kubeOptions.serviceOptions.Name = defaultOptions.AppName
		kubeOptions.serviceOptions.Namespace = kubeOptions.Namespace
		kubeOptions.serviceOptions.TargetPort = kubeOptions.deploymentOptions.Port
		if err := kube.CreateOrUpdateService(clientset, ctx, kubeOptions.serviceOptions); err != nil {
			panic(err)
		}

		kubeOptions.ingressOptions.Name = defaultOptions.AppName
		kubeOptions.ingressOptions.Namespace = kubeOptions.Namespace
		if err := kube.CreateOrUpdateIngress(clientset, ctx, kubeOptions.ingressOptions); err != nil {
			panic(err)
		}

		if kubeOptions.hpaOptions.Enabled {
			kubeOptions.hpaOptions.Name = defaultOptions.AppName
			kubeOptions.hpaOptions.Namespace = kubeOptions.Namespace
			if err := kube.CreateOrUpdateHPA(clientset, ctx, kubeOptions.hpaOptions); err != nil {
				panic(err)
			}
		} else {
			kubeOptions.hpaOptions.Name = defaultOptions.AppName
			kubeOptions.hpaOptions.Namespace = kubeOptions.Namespace
			if err := kube.DeleteHPA(clientset, ctx, kubeOptions.hpaOptions); err != nil {
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

	if helpers.IsBlank(kubeOptions.ingressOptions.Host) {
		kubeOptions.ingressOptions.Host = fmt.Sprintf("%s.com", defaultOptions.AppName)
	}
}
