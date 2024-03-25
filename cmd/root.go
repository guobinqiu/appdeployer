package cmd

import (
	"context"
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
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	// default
	rootCmd.Flags().StringVar(&defaultOptions.Kubeconfig, "default.kubeconfig", viper.GetString("default.kubeconfig"), "default.kubeconfig")
	rootCmd.Flags().StringVar(&defaultOptions.AppDir, "default.appdir", viper.GetString("default.kubeconfig"), "default.appdir")
	rootCmd.Flags().StringVar(&defaultOptions.ApplicationName, "default.applicationName", viper.GetString("default.applicationName"), "default.applicationName")
	rootCmd.Flags().StringVar(&defaultOptions.Namespace, "default.namespace", viper.GetString("default.namespace"), "default.namespace")

	// git
	rootCmd.Flags().BoolVar(&gitOptions.Pull, "git.pull", viper.GetBool("git.pull"), "git.pull")
	rootCmd.Flags().StringVar(&gitOptions.Repo, "git.repo", viper.GetString("git.repo"), "git.repo")
	rootCmd.Flags().StringVar(&gitOptions.Username, "git.username", viper.GetString("git.username"), "git.username")
	rootCmd.Flags().StringVar(&gitOptions.Password, "git.password", viper.GetString("git.password"), "git.password")

	// docker
	rootCmd.Flags().StringVar(&dockerOptions.Configfile, "docker.configfile", viper.GetString("docker.configfile"), "docker.configfile")
	rootCmd.Flags().StringVar(&dockerOptions.Dockerfile, "docker.dockerfile", viper.GetString("docker.dockerfile"), "docker.dockerfile")
	rootCmd.Flags().StringVar(&dockerOptions.Registry, "docker.registry", viper.GetString("docker.reregistry"), "docker.reregistry")
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
		//TODO
		ctx := context.TODO()

		appdir := defaultOptions.AppDir
		if helpers.IsBlank(appdir) {
			panic("appdir cannot be empty")
		}

		exist, err := helpers.IsDirExist(appdir)
		if err != nil {
			panic(err)
		}
		if !exist {
			panic("appdir does not exist")
		}

		gitOptions.AppDir = appdir
		dockerOptions.AppDir = appdir

		// Pull or clone to appdir
		if gitOptions.Pull {
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

		// Get kubeconfig file path
		kubeconfigPath := defaultOptions.Kubeconfig
		if helpers.IsBlank(kubeconfigPath) {
			kubeconfigPath = helpers.GetDefaultKubeconfigPath()
		}
		exist, err = helpers.IsFileExist(kubeconfigPath)
		if err != nil {
			panic(err)
		}
		if !exist {
			panic("kubeconfig does not exist")
		}

		// Create a kubernetes client by the specified kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			panic(err)
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}

		applicationName := defaultOptions.ApplicationName
		if helpers.IsBlank(applicationName) {
			applicationName = filepath.Base(appdir)
		}

		namespace := defaultOptions.Namespace
		if helpers.IsBlank(namespace) {
			namespace = applicationName
		}

		// Update (or create if not exists) kubernetes resource objects
		resources.CreateOrUpdateNamespace(clientset, ctx, namespace)

		resources.CreateOrUpdateDockerSecret(clientset, ctx, resources.DockerSecretOptions{
			ApplicationName: applicationName,
			Namespace:       namespace,
			DockerOptions:   dockerOptions,
		})

		resources.CreateOrUpdateServiceAccount(clientset, ctx, resources.ServiceAccountOptions{
			ApplicationName: applicationName,
			Namespace:       namespace,
		})

		deploymentOptions.ApplicationName = applicationName
		deploymentOptions.Namespace = namespace
		deploymentOptions.Image = dockerOptions.Image()
		resources.CreateOrUpdateDeployment(clientset, ctx, deploymentOptions)

		serviceOptions.ApplicationName = applicationName
		serviceOptions.Namespace = namespace
		serviceOptions.TargetPort = deploymentOptions.Port
		resources.CreateOrUpdateService(clientset, ctx, serviceOptions)

		ingressOptions.ApplicationName = applicationName
		ingressOptions.Namespace = namespace
		resources.CreateOrUpdateIngress(clientset, ctx, ingressOptions)
	},
}
