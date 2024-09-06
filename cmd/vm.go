package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/guobinqiu/appdeployer/git"
	"github.com/guobinqiu/appdeployer/helpers"
	"github.com/guobinqiu/appdeployer/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SSHOptions struct {
	Username              string `form:"username" json:"username"`
	Password              string `form:"password" json:"password"`
	Port                  int    `form:"port" json:"port"`
	AuthorizedKeysPath    string `form:"authorized_keys_path" json:"authorized_keys_path"`
	PrivatekeyPath        string `form:"privatekey_path" json:"privatekey_path"`
	PublickeyPath         string `form:"publickey_path" json:"publickey_path"`
	KnownHostsPath        string `form:"knownhosts_path" json:"knownhosts_path"`
	StrictHostKeyChecking bool   `form:"stricthostkeychecking" json:"stricthostkeychecking"`
}

type AnsibleOptions struct {
	Hosts          string `form:"hosts" json:"hosts"`
	Role           string `form:"role" json:"role"`
	BecomePassword string `form:"become_password" json:"become_password"`
	InstallDir     string `form:"installdir" json:"installdir"`
}

var sshOptions SSHOptions
var ansibleOptions AnsibleOptions

func init() {
	// set default values
	viper.SetDefault("ssh.port", 22)
	viper.SetDefault("ssh.authorized_keys_path", "~/.ssh/authorized_keys")
	viper.SetDefault("ssh.privatekey_path", "~/.ssh/appdeployer")
	viper.SetDefault("ssh.publickey_path", "~/.ssh/appdeployer.pub")
	viper.SetDefault("ssh.knownhosts_path", "~/.ssh/known_hosts")
	viper.SetDefault("ssh.stricthostkeychecking", true)
	viper.SetDefault("ansible.hosts", "localhost")
	viper.SetDefault("ansible.installdir", "~/workspace")

	//ssh
	vmCmd.Flags().StringVar(&sshOptions.Username, "ssh.username", viper.GetString("ssh.username"), "Username for connecting to SSH server")
	vmCmd.Flags().StringVar(&sshOptions.Password, "ssh.password", viper.GetString("ssh.password"), "Password for connecting to SSH server")
	vmCmd.Flags().IntVar(&sshOptions.Port, "ssh.port", viper.GetInt("ssh.port"), "Port for connecting to SSH server")
	vmCmd.Flags().StringVar(&sshOptions.AuthorizedKeysPath, "ssh.authorized_keys_path", viper.GetString("ssh.authorized_keys_path"), "Path to SSH server authorized_keys file storing SSH client's public keys. Defaults to ~/.ssh/authorized_keys")
	vmCmd.Flags().StringVar(&sshOptions.PrivatekeyPath, "ssh.privatekey_path", viper.GetString("ssh.privatekey_path"), "Path to SSH client private key file")
	vmCmd.Flags().StringVar(&sshOptions.PublickeyPath, "ssh.publickey_path", viper.GetString("ssh.publickey_path"), "Path to SSH client public key file")
	vmCmd.Flags().StringVar(&sshOptions.KnownHostsPath, "ssh.knownhosts_path", viper.GetString("ssh.knownhosts_path"), "Path to SSH client known_hosts file storing SSH server's public keys. Defaults to ~/.ssh/known_hosts")
	vmCmd.Flags().BoolVar(&sshOptions.StrictHostKeyChecking, "ssh.stricthostkeychecking", viper.GetBool("ssh.stricthostkeychecking"), "Whether or not to skip the confirmation of the SSH server's public key. Defaults to true")

	//ansible
	vmCmd.Flags().StringVar(&ansibleOptions.Hosts, "ansible.hosts", viper.GetString("ansible.hosts"), "Hosts on which the app will be deployed. Defaults to localhost.")
	vmCmd.Flags().StringVar(&ansibleOptions.Role, "ansible.role", viper.GetString("ansible.role"), "Run ansible playbook by role for your app. Such as go, java and nodejs")
	vmCmd.Flags().StringVar(&ansibleOptions.BecomePassword, "ansible.become_password", viper.GetString("ansible.become_password"), "Run ansible playbook with sudo privileges")
	vmCmd.Flags().StringVar(&ansibleOptions.InstallDir, "ansible.installdir", viper.GetString("ansible.installdir"), "Directory where the app will be installed. Defaults to ~/workspace")
}

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Deploy app to VM set",
	RunE: func(cmd *cobra.Command, args []string) error {
		return VMDeploy(&defaultOptions, &gitOptions, &sshOptions, &ansibleOptions, func(msg string) {
			fmt.Println(msg)
		})
	},
}

func VMDeploy(defaultOptions *DefaultOptions, gitOptions *git.GitOptions, sshOptions *SSHOptions, ansibleOptions *AnsibleOptions, logHandler func(msg string)) error {
	if err := setDefaultOptions(defaultOptions); err != nil {
		return err
	}

	if err := setSSHOptions(sshOptions); err != nil {
		return err
	}

	if err := setAnsibleOptions(ansibleOptions); err != nil {
		return err
	}

	if err := gitPull(gitOptions, logHandler); err != nil {
		return err
	}

	if err := setupAnsible(sshOptions, ansibleOptions); err != nil {
		return err
	}

	if err := runPlaybook(defaultOptions, sshOptions, ansibleOptions); err != nil {
		return err
	}

	return nil
}

func setSSHOptions(sshOptions *SSHOptions) error {
	if helpers.IsBlank(sshOptions.Username) {
		return fmt.Errorf("ssh.username is required")
	}

	if helpers.IsBlank(sshOptions.Password) {
		return fmt.Errorf("ssh.password is required")
	}

	return nil
}

func setAnsibleOptions(ansibleOptions *AnsibleOptions) error {
	if helpers.IsBlank(ansibleOptions.Role) {
		return fmt.Errorf("ansible.role is required")
	}

	currentDir, _ := os.Getwd()
	rootDir := helpers.FindRootDir(currentDir)
	roles, err := helpers.ListSubDirs(filepath.Join(rootDir, "ansible_roles/"))
	if err != nil {
		return err
	}

	if !helpers.Contains(roles, ansibleOptions.Role) {
		return fmt.Errorf("role should be one of %s", roles)
	}

	if helpers.IsBlank(ansibleOptions.BecomePassword) {
		return fmt.Errorf("ansible.become_password is required")
	}

	return nil
}

func setupAnsible(sshOptions *SSHOptions, ansibleOptions *AnsibleOptions) error {
	keyManager := ssh.NewSSHKeyManager(
		ssh.WithPrivateKeyPath(sshOptions.PrivatekeyPath),
		ssh.WithPublicKeyPath(sshOptions.PublickeyPath),
		ssh.WithKnownHostsPath(sshOptions.KnownHostsPath),
		ssh.WithTimeout(10*time.Second),
		ssh.WithStrictHostKeyChecking(sshOptions.StrictHostKeyChecking),
	)
	hosts := strings.Split(ansibleOptions.Hosts, ",")
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		keyfile := sshOptions.PrivatekeyPath
		keyfileExist, err := helpers.IsFileExist(keyfile)
		if err != nil {
			return fmt.Errorf("failed to check the existence of the SSH key file (%s): %v", keyfile, err)
		}
		if !keyfileExist {
			if err := keyManager.GenerateAndSaveKeyPair(); err != nil {
				return fmt.Errorf("failed to generate and save SSH key pair: %v", err)
			}
		}
		if err := keyManager.AddPublicKeyToRemote(host, sshOptions.Port, sshOptions.Username, sshOptions.Password, sshOptions.AuthorizedKeysPath); err != nil {
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

func runPlaybook(defaultOptions *DefaultOptions, sshOptions *SSHOptions, ansibleOptions *AnsibleOptions) error {
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
		"-u", sshOptions.Username,
		"-e", fmt.Sprintf("ansible_port=%d", sshOptions.Port),
		"-e", fmt.Sprintf("ansible_ssh_private_key_file=%s", sshOptions.PrivatekeyPath),
		"-e", fmt.Sprintf("ansible_become_password=%s", ansibleOptions.BecomePassword),
		playbookFile.Name(),
	}

	cmd := exec.Command("ansible-playbook", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	currentDir, _ := os.Getwd()
	rootDir := helpers.FindRootDir(currentDir)
	cmd.Dir = filepath.Join(rootDir)

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
