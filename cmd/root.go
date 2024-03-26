package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/guobinqiu/deployer/docker"
	"github.com/guobinqiu/deployer/git"
	"github.com/guobinqiu/deployer/helpers"
	"github.com/guobinqiu/deployer/resources"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type DefaultOptions struct {
	Kubeconfig      string
	AppDir          string
	ApplicationName string
	Namespace       string
}

var (
	defaultOptions    DefaultOptions
	gitOptions        git.GitOptions
	dockerOptions     docker.DockerOptions
	ingressOptions    resources.IngressOptions
	serviceOptions    resources.ServiceOptions
	deploymentOptions resources.DeploymentOptions
)

func init() {
	viper.SetConfigFile("./config.ini")
	viper.ReadInConfig()

	// set default values
	viper.SetDefault("default.kubeconfig", helpers.GetDefaultKubeConfig())
	viper.SetDefault("git.pull", false)
	viper.SetDefault("docker.dockerconfig", helpers.GetDefaultDockerConfig())
	viper.SetDefault("docker.dockerfile", "./Dockerfile")
	viper.SetDefault("docker.registry", docker.DOCKERHUB)
	viper.SetDefault("docker.tag", "latest")
	viper.SetDefault("ingress.tls", true)
	viper.SetDefault("service.port", 8000)
	viper.SetDefault("deployment.replicas", 1)
	viper.SetDefault("deployment.port", 8000)

	// default
	rootCmd.Flags().StringVar(&defaultOptions.Kubeconfig, "default.kubeconfig", viper.GetString("default.kubeconfig"), "default.kubeconfig")
	rootCmd.Flags().StringVar(&defaultOptions.AppDir, "default.appdir", viper.GetString("default.appdir"), "default.appdir")
	rootCmd.Flags().StringVar(&defaultOptions.ApplicationName, "default.applicationName", viper.GetString("default.applicationName"), "default.applicationName")
	rootCmd.Flags().StringVar(&defaultOptions.Namespace, "default.namespace", viper.GetString("default.namespace"), "default.namespace")

	// git
	rootCmd.Flags().BoolVar(&gitOptions.Pull, "git.pull", viper.GetBool("git.pull"), "git.pull")
	rootCmd.Flags().StringVar(&gitOptions.Repo, "git.repo", viper.GetString("git.repo"), "git.repo")
	rootCmd.Flags().StringVar(&gitOptions.Username, "git.username", viper.GetString("git.username"), "git.username")
	rootCmd.Flags().StringVar(&gitOptions.Password, "git.password", viper.GetString("git.password"), "git.password")

	// docker
	rootCmd.Flags().StringVar(&dockerOptions.Dockerconfig, "docker.dockerconfig", viper.GetString("docker.dockerconfig"), "docker.dockerconfig")
	rootCmd.Flags().StringVar(&dockerOptions.Dockerfile, "docker.dockerfile", viper.GetString("docker.dockerfile"), "docker.dockerfile")
	rootCmd.Flags().StringVar(&dockerOptions.Registry, "docker.registry", viper.GetString("docker.registry"), "docker.registry")
	rootCmd.Flags().StringVar(&dockerOptions.Username, "docker.username", viper.GetString("docker.username"), "docker.username")
	rootCmd.Flags().StringVar(&dockerOptions.Password, "docker.password", viper.GetString("docker.password"), "docker.password")
	rootCmd.Flags().StringVar(&dockerOptions.Repository, "docker.repository", viper.GetString("docker.repository"), "docker.repository")
	rootCmd.Flags().StringVar(&dockerOptions.Tag, "docker.tag", viper.GetString("docker.tag"), "docker.tag")

	// ingress
	rootCmd.Flags().StringVar(&ingressOptions.Host, "ingress.host", viper.GetString("ingress.host"), "ingress.host")
	rootCmd.Flags().BoolVar(&ingressOptions.TLS, "ingress.tls", viper.GetBool("ingress.tls"), "ingress.tls")

	// service
	rootCmd.Flags().Int32Var(&serviceOptions.Port, "service.port", viper.GetInt32("service.port"), "service.port")

	// deployment
	rootCmd.Flags().Int32Var(&deploymentOptions.Replicas, "deployment.replicas", viper.GetInt32("deployment.replicas"), "deployment.replicas")
	rootCmd.Flags().Int32Var(&deploymentOptions.Port, "deployment.port", viper.GetInt32("deployment.port"), "deployment.port")
}

func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "deploy",
	Short: "app deployer",
	Long:  "app deployer is used to deploy your application to any kubernetes clusters as well as VMs via ansible",
	Run: func(cmd *cobra.Command, args []string) {
		setDefault()
		print()

		//TODO handle timeout or cancel
		ctx := context.TODO()

		// Pull or clone to appdir
		if gitOptions.Pull {
			if helpers.IsBlank(gitOptions.Repo) {
				panic("git.repo is required")
			}
			if err := git.Pull(gitOptions); err != nil {
				panic(err)
			}
		}

		// Create a docker service
		dockerservice, err := docker.NewDockerService()
		if err != nil {
			panic(err)
		}

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
		config, err := clientcmd.BuildConfigFromFlags("", defaultOptions.Kubeconfig)
		if err != nil {
			panic(err)
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}

		// Update (or create if not exists) kubernetes resource objects
		resources.CreateOrUpdateNamespace(clientset, ctx, defaultOptions.Namespace)

		resources.CreateOrUpdateDockerSecret(clientset, ctx, resources.DockerSecretOptions{
			ApplicationName: defaultOptions.ApplicationName,
			Namespace:       defaultOptions.Namespace,
			DockerOptions:   dockerOptions,
		})

		resources.CreateOrUpdateServiceAccount(clientset, ctx, resources.ServiceAccountOptions{
			ApplicationName: defaultOptions.ApplicationName,
			Namespace:       defaultOptions.Namespace,
		})

		img, _ := dockerOptions.Image()
		deploymentOptions.ApplicationName = defaultOptions.ApplicationName
		deploymentOptions.Namespace = defaultOptions.Namespace
		deploymentOptions.Image = img
		resources.CreateOrUpdateDeployment(clientset, ctx, deploymentOptions)

		serviceOptions.ApplicationName = defaultOptions.ApplicationName
		serviceOptions.Namespace = defaultOptions.Namespace
		serviceOptions.TargetPort = deploymentOptions.Port
		resources.CreateOrUpdateService(clientset, ctx, serviceOptions)

		ingressOptions.ApplicationName = defaultOptions.ApplicationName
		ingressOptions.Namespace = defaultOptions.Namespace
		resources.CreateOrUpdateIngress(clientset, ctx, ingressOptions)
	},
}

func setDefault() {
	defaultOptions.AppDir = helpers.ExpandUser(defaultOptions.AppDir)
	if helpers.IsBlank(defaultOptions.AppDir) {
		panic("appdir is required")
	}

	exist, err := helpers.IsDirExist(defaultOptions.AppDir)
	if err != nil {
		panic(err)
	}
	if !exist {
		panic("appdir does not exist")
	}
	gitOptions.AppDir = defaultOptions.AppDir
	dockerOptions.AppDir = defaultOptions.AppDir

	defaultOptions.Kubeconfig = helpers.ExpandUser(defaultOptions.Kubeconfig)
	exist, err = helpers.IsFileExist(defaultOptions.Kubeconfig)
	if err != nil {
		panic(err)
	}
	if !exist {
		panic("kubeconfig does not exist")
	}

	dockerOptions.Dockerconfig = helpers.ExpandUser(dockerOptions.Dockerconfig)
	exist, err = helpers.IsFileExist(dockerOptions.Dockerconfig)
	if err != nil {
		panic(err)
	}
	if !exist {
		panic("dockerconfig does not exist")
	}

	applicationName := filepath.Base(defaultOptions.AppDir)

	if helpers.IsBlank(defaultOptions.ApplicationName) {
		defaultOptions.ApplicationName = applicationName
	}

	if helpers.IsBlank(defaultOptions.Namespace) {
		defaultOptions.Namespace = applicationName
	}

	if helpers.IsBlank(ingressOptions.Host) {
		ingressOptions.Host = fmt.Sprintf("%s.com", applicationName)
	}

	if helpers.IsBlank(dockerOptions.Repository) && dockerOptions.Registry == docker.DOCKERHUB {
		if helpers.IsBlank(dockerOptions.Username) {
			panic("docker.username is required")
		}
		dockerOptions.Repository = fmt.Sprintf("%s/%s", dockerOptions.Username, applicationName)
	}
}

func print() {
	fmt.Printf("%+v\n", defaultOptions)
	fmt.Printf("%+v\n", gitOptions)
	fmt.Printf("%+v\n", dockerOptions)
	fmt.Printf("%+v\n", ingressOptions)
	fmt.Printf("%+v\n", serviceOptions)
	fmt.Printf("%+v\n", deploymentOptions)
}
