package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/guobinqiu/appdeployer/git"
	"github.com/guobinqiu/appdeployer/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type DefaultOptions struct {
	AppDir  string `form:"appdir" json:"appdir"`
	AppName string `form:"appname" json:"appname"`
}

var (
	defaultOptions DefaultOptions
	gitOptions     git.GitOptions
)

func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "appdeploy",
	Short: "App deployer",
	Long:  "App deployer is used to deploy your application to any kubernetes clusters as well as VMs via ansible",
}

func init() {
	viper.SetConfigFile("./config.ini")
	viper.ReadInConfig()

	// default
	rootCmd.PersistentFlags().StringVar(&defaultOptions.AppDir, "default.appdir", viper.GetString("default.appdir"), "App installation directory")
	rootCmd.PersistentFlags().StringVar(&defaultOptions.AppName, "default.appname", viper.GetString("default.appname"), "Name of app. Defaults to name of app installation directory")

	// git
	rootCmd.Flags().BoolVar(&gitOptions.Enabled, "git.enabled", viper.GetBool("git.enabled"), "Enable or disable git pull")
	rootCmd.Flags().StringVar(&gitOptions.Repo, "git.repo", viper.GetString("git.repo"), "URL of git repository")
	rootCmd.Flags().StringVar(&gitOptions.Username, "git.username", viper.GetString("git.username"), "Username for git")
	rootCmd.Flags().StringVar(&gitOptions.Password, "git.password", viper.GetString("git.password"), "Password for git")

	// Add sub commands to root command
	rootCmd.AddCommand(kubeCmd)
	rootCmd.AddCommand(vmCmd)
}

func setDefaultOptions(defaultOptions *DefaultOptions) error {
	defaultOptions.AppDir = helpers.ExpandUser(defaultOptions.AppDir)
	if helpers.IsBlank(defaultOptions.AppDir) {
		return fmt.Errorf("appdir is required")
	}

	exist, err := helpers.IsDirExist(defaultOptions.AppDir)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("appdir does not exist")
	}

	if helpers.IsBlank(defaultOptions.AppName) {
		defaultOptions.AppName = filepath.Base(defaultOptions.AppDir)
	}

	return nil
}

// Pull or clone into appdir
func gitPull(gitOptions *git.GitOptions, logHandler func(msg string)) error {
	gitOptions.AppDir = defaultOptions.AppDir
	if gitOptions.Enabled {
		if helpers.IsBlank(gitOptions.Repo) {
			return fmt.Errorf("git.repo is required")
		}
		if err := git.Pull(*gitOptions, logHandler); err != nil {
			return err
		}
	}

	return nil
}
