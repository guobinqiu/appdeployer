package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/guobinqiu/deployer/ansible"
	"github.com/guobinqiu/deployer/git"
	"github.com/guobinqiu/deployer/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"text/template"
)

type SSHOptions struct {
	Username                 string
	Password                 string
	Port                     int
	HomeDir                  string
	ClientKeyFileName        string
	ServerAuthorizedKeysPath string
}

type AnsibleOptions struct {
	Hosts                    string
	Inventory                string
	Role                     string
	AnsibleUser              string
	AnsiblePort              int
	AnsibleSSHPrivateKeyFile string
	AnsibleBecomePassword    string
}

var sshOptions SSHOptions
var ansibleOptions AnsibleOptions

func init() {
	// set default values
	viper.SetDefault("ssh.port", 22)
	viper.SetDefault("ssh.homedir", os.Getenv("HOME"))
	viper.SetDefault("ssh.client_keyfilename", "deployer")
	viper.SetDefault("ssh.server_authorized_keys_path", "~/.ssh/authorized_keys")
	viper.SetDefault("ansible.ansible_port", 22)
	viper.SetDefault("ansible.ansible_ssh_private_key_file", "~/.ssh/id_rsa")

	//ssh
	vmCmd.Flags().StringVar(&sshOptions.Username, "ssh.username", viper.GetString("ssh.username"), "ssh.username")
	vmCmd.Flags().StringVar(&sshOptions.Password, "ssh.password", viper.GetString("ssh.password"), "ssh.password")
	vmCmd.Flags().StringVar(&sshOptions.HomeDir, "ssh.homedir", viper.GetString("ssh.homedir"), "ssh.homedir")
	vmCmd.Flags().StringVar(&sshOptions.ClientKeyFileName, "ssh.client_keyfilename", viper.GetString("ssh.client_keyfilename"), "ssh.client_keyfilename")
	vmCmd.Flags().StringVar(&sshOptions.ServerAuthorizedKeysPath, "ssh.server_authorized_keys_path", viper.GetString("ssh.server_authorized_keys_path"), "ssh.server_authorized_keys_path")

	//ansible
	vmCmd.Flags().StringVar(&ansibleOptions.Hosts, "ansible.hosts", viper.GetString("ansible.hosts"), "ansible.hosts")
	vmCmd.Flags().StringVar(&ansibleOptions.Role, "ansible.role", viper.GetString("ansible.role"), "ansible.role")
	vmCmd.Flags().StringVar(&ansibleOptions.AnsibleUser, "ansible.ansible_user", viper.GetString("ansible.ansible_user"), "ansible.ansible_user")
	vmCmd.Flags().IntVar(&ansibleOptions.AnsiblePort, "ansible.ansible_port", viper.GetInt("ansible.ansible_port"), "ansible.ansible_port")
	vmCmd.Flags().StringVar(&ansibleOptions.AnsibleSSHPrivateKeyFile, "ansible.ansible_ssh_private_key_file", viper.GetString("ansible.ansible_ssh_private_key_file"), "ansible.ansible_ssh_private_key_file")
	vmCmd.Flags().StringVar(&ansibleOptions.AnsibleBecomePassword, "ansible.ansible_become_password", viper.GetString("ansible.ansible_become_password"), "ansible.ansible_become_password")
}

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Deploy to VM",
	Run: func(cmd *cobra.Command, args []string) {
		setDefaultOptions()
		setSSHOptions()
		setAnsibleOptions()

		// Pull or clone into appdir
		if gitOptions.Pull {
			if helpers.IsBlank(gitOptions.Repo) {
				panic("git.repo is required")
			}
			if err := git.Pull(gitOptions); err != nil {
				panic(err)
			}
		}

		if err := setupAnsible(ansibleOptions, sshOptions); err != nil {
			panic(err)
		}

		if err := runPlaybook(); err != nil {
			panic(err)
		}
	},
}

func setSSHOptions() {
	if helpers.IsBlank(sshOptions.Username) {
		panic("ssh.username is required")
	}

	if helpers.IsBlank(sshOptions.Password) {
		panic("ssh.password is required")
	}
}

func setAnsibleOptions() {
	if helpers.IsBlank(ansibleOptions.Hosts) {
		panic("ansible.hosts is required")
	}

	if helpers.IsBlank(ansibleOptions.AnsibleBecomePassword) {
		panic("ansible.ansible_become_password is required")
	}

	ansibleOptions.AnsibleUser = sshOptions.Username

	//check hosts

	//check role in ["go", "java", "nodejs"]
}

func setupAnsible(ansibleOptions AnsibleOptions, sshOptions SSHOptions) error {
	hosts := strings.Split(ansibleOptions.Hosts, ",")
	for _, host := range hosts {
		keyManager := ansible.NewSSHKeyManager(
			ansible.WithHomeDir(sshOptions.HomeDir),
			ansible.WithKeyFileName(sshOptions.ClientKeyFileName),
		)
		if err := keyManager.GenerateAndSaveKeyPair(); err != nil {
			return fmt.Errorf("failed to generate and save SSH key pair: %v", err)
		}
		if err := keyManager.AddPublicKeyToRemote(host, ansibleOptions.AnsiblePort, sshOptions.Username, sshOptions.Password, sshOptions.ServerAuthorizedKeysPath); err != nil {
			return fmt.Errorf("failed to add public key to remote host %s: %v", host, err)
		}
	}
	return nil
}

const playbookTemplate = `
---
- hosts: {{ .Hosts }}
  gather_facts: yes
  become: yes
  vars:
    appdir: {{ .AppDir }}
  roles:
    - role: {{ .Role }}
`

type PlaybookData struct {
	Hosts  string
	AppDir string
	Role   string
}

func runPlaybook() error {
	tmpl, err := template.New("playbook").Parse(playbookTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse playbook template: %v", err)
	}

	var playbook bytes.Buffer

	if err := tmpl.Execute(&playbook, PlaybookData{
		Hosts:  ansibleOptions.Hosts,
		AppDir: defaultOptions.AppDir,
		Role:   ansibleOptions.Role,
	}); err != nil {
		return fmt.Errorf("failed to execute playbook template: %v", err)
	}

	tempFile, err := os.CreateTemp("/tmp", "playbook-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(playbook.Bytes()); err != nil {
		return fmt.Errorf("failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	cmd := "ansible-playbook"
	args := []string{
		"-u", ansibleOptions.AnsibleUser,
		"-e", fmt.Sprintf("ansible_ssh_private_key_file=%s", ansibleOptions.AnsibleSSHPrivateKeyFile),
		"-e", fmt.Sprintf("ansible_become_password=%s", ansibleOptions.AnsibleBecomePassword),
		tempFile.Name(),
	}

	output, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return fmt.Errorf("failed to execute playbook: %v", err)
	}

	fmt.Println(string(output))
	return nil
}
