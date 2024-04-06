package cmd

import (
	"path/filepath"

	"github.com/guobinqiu/deployer/git"
	"github.com/guobinqiu/deployer/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type DefaultOptions struct {
	AppDir  string
	AppName string
}

var (
	defaultOptions DefaultOptions
	gitOptions     git.GitOptions
)

func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "deploy",
	Short: "App deployer",
	Long:  "App deployer is used to deploy your application to any kubernetes clusters as well as VMs via ansible",
}

func init() {
	viper.SetConfigFile("./config.ini")
	viper.ReadInConfig()

	// default
	rootCmd.PersistentFlags().StringVar(&defaultOptions.AppDir, "default.appdir", viper.GetString("default.appdir"), "default.appdir")
	rootCmd.PersistentFlags().StringVar(&defaultOptions.AppName, "default.appname", viper.GetString("default.appname"), "default.appname")

	// git
	rootCmd.Flags().BoolVar(&gitOptions.Pull, "git.pull", viper.GetBool("git.pull"), "git.pull")
	rootCmd.Flags().StringVar(&gitOptions.Repo, "git.repo", viper.GetString("git.repo"), "git.repo")
	rootCmd.Flags().StringVar(&gitOptions.Username, "git.username", viper.GetString("git.username"), "git.username")
	rootCmd.Flags().StringVar(&gitOptions.Password, "git.password", viper.GetString("git.password"), "git.password")

	// Add sub commands to root command
	rootCmd.AddCommand(kubeCmd)
	rootCmd.AddCommand(vmCmd)
}

func setDefaultOptions() {
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

	if helpers.IsBlank(defaultOptions.AppName) {
		defaultOptions.AppName = filepath.Base(defaultOptions.AppDir)
	}
}

// Pull or clone into appdir
func gitPull() {
	gitOptions.AppDir = defaultOptions.AppDir
	if gitOptions.Pull {
		if helpers.IsBlank(gitOptions.Repo) {
			panic("git.repo is required")
		}
		if err := git.Pull(gitOptions); err != nil {
			panic(err)
		}
	}
}
