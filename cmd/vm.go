package cmd

import (
	"os"

	"github.com/guobinqiu/deployer/ansible"
	"github.com/guobinqiu/deployer/git"
	"github.com/guobinqiu/deployer/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ansibleOptions = struct {
	Hosts             string
	Inventory         string
	Role              string
	User              string
	Port              int32
	SSHPrivateKeyFile string
	BecomePassword    string
}{}

func init() {
	//ansible
	kubeCmd.Flags().StringVar(&ansibleOptions.Hosts, "ansible.hosts", viper.GetString("ansible.hosts"), "ansible.hosts")
	kubeCmd.Flags().StringVar(&ansibleOptions.Inventory, "ansible.inventory", viper.GetString("ansible.inventory"), "ansible.inventory")
	kubeCmd.Flags().StringVar(&ansibleOptions.Role, "ansible.role", viper.GetString("ansible.role"), "ansible.role")
	kubeCmd.Flags().StringVar(&ansibleOptions.User, "ansible.user", viper.GetString("ansible.user"), "ansible.user")
	kubeCmd.Flags().Int32Var(&ansibleOptions.Port, "ansible.port", viper.GetInt32("ansible.port"), "ansible.port")
	kubeCmd.Flags().StringVar(&ansibleOptions.SSHPrivateKeyFile, "ansible.ssh_private_key_file", viper.GetString("ansible.ssh_private_key_file"), "ansible.ssh_private_key_file")
	kubeCmd.Flags().StringVar(&ansibleOptions.BecomePassword, "ansible.become_password", viper.GetString("ansible.become_password"), "ansible.become_password")
}

var VMCmd = &cobra.Command{
	Use:   "vm",
	Short: "Deploy to VM",
	Run: func(cmd *cobra.Command, args []string) {
		//setDefaultOptions()

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

		runPlaybook()
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

func runPlaybook() {

}
