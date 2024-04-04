package cmd

import (
	"os"

	"github.com/guobinqiu/deployer/ansible"
	"github.com/guobinqiu/deployer/git"
	"github.com/guobinqiu/deployer/helpers"
	"github.com/spf13/cobra"
)

var VMCmd = &cobra.Command{
	Use:   "vm",
	Short: "Deploy to VM",
	Run: func(cmd *cobra.Command, args []string) {
		// Pull or clone into appdir
		if gitOptions.Pull {
			if helpers.IsBlank(gitOptions.Repo) {
				panic("git.repo is required")
			}
			if err := git.Pull(gitOptions); err != nil {
				panic(err)
			}
		}

		setupAnsible()

	},
}

func setupAnsible() {
	homeDir := os.Getenv("HOME")
	keyManager := ansible.NewSSHKeyManager(
		ansible.WithHomeDir(homeDir),
		ansible.WithKeyFileName("deployer"),
	)
	if err := keyManager.GenerateAndSaveKeyPair(); err != nil {
		panic(err)
	}

	host := "192.168.1.9"
	port := 22
	username := "guobin"
	password := "111111"
	if err := keyManager.AddPublicKeyToRemote(host, port, username, password); err != nil {
		panic(err)
	}
}
