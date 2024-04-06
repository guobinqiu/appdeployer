package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/guobinqiu/deployer/ansible"
	"github.com/guobinqiu/deployer/git"
	"github.com/guobinqiu/deployer/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// go run main.go vm --default.appdir=~/workspace/hellojava --ssh.username=guobin --ssh.password=111111 --ansible.ansible_become_password=111111 --ansible.hosts=127.0.0.1 --ansible.ansible_port=2222 --ansible.role=java
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

		// if err := setupAnsible(ansibleOptions, sshOptions); err != nil {
		// 	panic(err)
		// }

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
	ansibleOptions.AnsibleUser = sshOptions.Username

	//check hosts
	if helpers.IsBlank(ansibleOptions.Hosts) {
		panic("ansible.hosts is required")
	}
	//...

	//check role in ["go", "java", "nodejs"]
	if helpers.IsBlank(ansibleOptions.Role) {
		panic("ansible.role is required")
	}
	//...

	if helpers.IsBlank(ansibleOptions.AnsibleBecomePassword) {
		panic("ansible.ansible_become_password is required")
	}
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

const inventoryTemplate = `
[{{ .GroupName }}]
{{ .Hosts }}
`

const playbookTemplate = `
---
- hosts: {{ .GroupName }}
  gather_facts: yes
  become: yes
  vars:
    appdir: {{ .AppDir }}
  roles:
    - role: {{ .Role }}
`

type InventoryData struct {
	GroupName string
	Hosts     string
}

type PlaybookData struct {
	GroupName string
	AppDir    string
	Role      string
}

func runPlaybook() error {
	tmpl, err := template.New("inventory").Parse(inventoryTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse inventory template: %v", err)
	}
	var inventory bytes.Buffer
	if err := tmpl.Execute(&inventory, InventoryData{
		GroupName: defaultOptions.ApplicationName,
		Hosts:     ansibleOptions.Hosts,
	}); err != nil {
		return fmt.Errorf("failed to execute inventory template: %v", err)
	}
	inventoryTempFile, err := os.CreateTemp("/tmp", "inventory-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create inventory temporary file: %v", err)
	}
	defer os.Remove(inventoryTempFile.Name())
	if _, err := inventoryTempFile.Write(inventory.Bytes()); err != nil {
		return fmt.Errorf("failed to write to inventory temporary file: %v", err)
	}
	inventoryTempFile.Close()

	tmpl, err = template.New("playbook").Parse(playbookTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse playbook template: %v", err)
	}
	var playbook bytes.Buffer
	if err := tmpl.Execute(&playbook, PlaybookData{
		GroupName: defaultOptions.ApplicationName,
		AppDir:    defaultOptions.AppDir,
		Role:      ansibleOptions.Role,
	}); err != nil {
		return fmt.Errorf("failed to execute playbook template: %v", err)
	}
	playbookTempFile, err := os.CreateTemp("/tmp", "playbook-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create playbook temporary file: %v", err)
	}
	defer os.Remove(playbookTempFile.Name())
	if _, err := playbookTempFile.Write(playbook.Bytes()); err != nil {
		return fmt.Errorf("failed to write to playbook temporary file: %v", err)
	}
	playbookTempFile.Close()

	cmd := "ansible-playbook"
	args := []string{
		"-i", inventoryTempFile.Name(),
		"-u", ansibleOptions.AnsibleUser,
		"-e", fmt.Sprintf("ansible_port=%d", ansibleOptions.AnsiblePort),
		"-e", fmt.Sprintf("ansible_ssh_private_key_file=%s", ansibleOptions.AnsibleSSHPrivateKeyFile),
		"-e", fmt.Sprintf("ansible_become_password=%s", ansibleOptions.AnsibleBecomePassword),
		playbookTempFile.Name(),
	}
	fmt.Println(args)

	command := exec.Command(cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
