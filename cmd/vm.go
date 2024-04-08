package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/guobinqiu/deployer/ansible"
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
	Role                     string
	AnsibleUser              string
	AnsiblePort              int
	AnsibleSSHPrivateKeyFile string
	AnsibleBecomePassword    string
	InstallDir               string
}

var sshOptions SSHOptions
var ansibleOptions AnsibleOptions

func init() {
	// set default values
	viper.SetDefault("ssh.port", 22)
	viper.SetDefault("ssh.homedir", os.Getenv("HOME"))
	viper.SetDefault("ssh.client_keyfilename", "deployer")
	viper.SetDefault("ssh.server_authorized_keys_path", "~/.ssh/authorized_keys")
	viper.SetDefault("ansible.hosts", "localhost")
	viper.SetDefault("ansible.ansible_port", 22)
	viper.SetDefault("ansible.ansible_ssh_private_key_file", "~/.ssh/deployer")
	viper.SetDefault("ansible.installdir", "~/workspace")

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
	vmCmd.Flags().StringVar(&ansibleOptions.InstallDir, "ansible.installdir", viper.GetString("ansible.installdir"), "ansible.installdir")
}

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Deploy to VM",
	Run: func(cmd *cobra.Command, args []string) {
		setDefaultOptions()
		setSSHOptions()
		setAnsibleOptions()

		gitPull()

		if err := setupAnsible(); err != nil {
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
	ansibleOptions.AnsibleUser = sshOptions.Username

	if helpers.IsBlank(ansibleOptions.Role) {
		panic("ansible.role is required")
	}

	roles, err := helpers.ListSubDirs("ansible_roles/")
	if err != nil {
		panic(err)
	}
	if !helpers.Contains(roles, ansibleOptions.Role) {
		panic(fmt.Sprintf("role should be one of %s", roles))
	}

	if helpers.IsBlank(ansibleOptions.AnsibleBecomePassword) {
		panic("ansible.ansible_become_password is required")
	}
}

func setupAnsible() error {
	hosts := strings.Split(ansibleOptions.Hosts, ",")
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		keyManager := ansible.NewSSHKeyManager(
			ansible.WithHomeDir(sshOptions.HomeDir),
			ansible.WithKeyFileName(sshOptions.ClientKeyFileName),
		)
		keyfile := helpers.ExpandUser(ansibleOptions.AnsibleSSHPrivateKeyFile)
		keyfileExist, err := helpers.IsFileExist(keyfile)
		if err != nil {
			return fmt.Errorf("failed to check the existence of the SSH key file (%s): %v", keyfile, err)
		}
		if !keyfileExist {
			if err := keyManager.GenerateAndSaveKeyPair(); err != nil {
				return fmt.Errorf("failed to generate and save SSH key pair: %v", err)
			}
		}
		if err := keyManager.AddPublicKeyToRemote(host, ansibleOptions.AnsiblePort, sshOptions.Username, sshOptions.Password, sshOptions.ServerAuthorizedKeysPath); err != nil {
			return fmt.Errorf("failed to add public key to remote host %s: %v", host, err)
		}
	}
	return nil
}

const inventoryTemplate = `
[{{ .AppName }}]
{{ .Hosts }}
`

const playbookTemplate = `
---
- hosts: {{ .AppName }}
  gather_facts: yes
  become: no
  vars:
    app_name: {{ .AppName }}
    app_dir: {{ .AppDir }}
    app_install_dir: {{ .InstallDir }}
  roles:
    - role: {{ .Role }}
`

type InventoryData struct {
	AppName string
	Hosts   string
}

type PlaybookData struct {
	AppName    string
	AppDir     string
	InstallDir string
	Role       string
}

func runPlaybook() error {
	inventoryFile, err := executeTemplate(inventoryTemplate, InventoryData{
		AppName: defaultOptions.AppName,
		Hosts:   ansibleOptions.Hosts,
	}, "inventory-*.yaml")
	if err != nil {
		return err
	}

	playbookFile, err := executeTemplate(playbookTemplate, PlaybookData{
		AppName:    defaultOptions.AppName,
		AppDir:     defaultOptions.AppDir,
		InstallDir: ansibleOptions.InstallDir,
		Role:       ansibleOptions.Role,
	}, "playbook-*.yaml")
	if err != nil {
		return err
	}

	cmdArgs := []string{
		"-i", inventoryFile.Name(),
		"-u", ansibleOptions.AnsibleUser,
		"-e", fmt.Sprintf("ansible_port=%d", ansibleOptions.AnsiblePort),
		"-e", fmt.Sprintf("ansible_ssh_private_key_file=%s", ansibleOptions.AnsibleSSHPrivateKeyFile),
		"-e", fmt.Sprintf("ansible_become_password=%s", ansibleOptions.AnsibleBecomePassword),
		playbookFile.Name(),
	}

	cmd := exec.Command("ansible-playbook", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute playbook: %v", err)
	}

	defer func() {
		inventoryFile.Close()
		os.Remove(inventoryFile.Name())

		playbookFile.Close()
		os.Remove(playbookFile.Name())
	}()

	return nil
}

func executeTemplate(templateStr string, data interface{}, filenamePattern string) (*os.File, error) {
	tmpl, err := template.New(filenamePattern[:strings.Index(filenamePattern, "*")]).Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %v", err)
	}

	tmpFile, err := os.CreateTemp("/tmp", filenamePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}

	writer := bufio.NewWriter(tmpFile)
	if err := tmpl.Execute(writer, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %v", err)
	}
	if err := writer.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush buffer to temporary file: %v", err)
	}

	return tmpFile, nil
}
