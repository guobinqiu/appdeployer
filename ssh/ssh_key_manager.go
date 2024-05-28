package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/guobinqiu/appdeployer/helpers"
	"golang.org/x/crypto/ssh"
)

type SSHKeyManager struct {
	KeyBitSize     int
	PrivateKeyPath string
	PublicKeyPath  string
}

func NewDefaultSSHKeyManager() *SSHKeyManager {
	return &SSHKeyManager{
		KeyBitSize:     2048,
		PrivateKeyPath: helpers.ExpandUser("~/.ssh/appdeployer"),
		PublicKeyPath:  helpers.ExpandUser("~/.ssh/appdeployer.pub"),
	}
}

func NewSSHKeyManager(options ...func(*SSHKeyManager)) *SSHKeyManager {
	manager := NewDefaultSSHKeyManager()
	for _, option := range options {
		option(manager)
	}
	return manager
}

func WithKeyBitSize(size int) func(*SSHKeyManager) {
	return func(m *SSHKeyManager) {
		m.KeyBitSize = size
	}
}

func WithPrivateKeyPath(path string) func(*SSHKeyManager) {
	return func(m *SSHKeyManager) {
		m.PrivateKeyPath = path
	}
}

func WithPublicKeyPath(path string) func(*SSHKeyManager) {
	return func(m *SSHKeyManager) {
		m.PublicKeyPath = path
	}
}

func (m *SSHKeyManager) GenerateAndSaveKeyPair() error {
	privateKey, publicKey, err := m.generateRSAKeyPair()
	if err != nil {
		return err
	}

	if err := m.savePrivateKey(privateKey); err != nil {
		return err
	}

	if err := m.savePublicKey(publicKey); err != nil {
		return err
	}

	return nil
}

func (m *SSHKeyManager) AddPublicKeyToRemote(host string, port int, username string, password string, remoteAuthorizedKeysPath string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return fmt.Errorf("failed to connect to the remote server: %w", err)
	}
	defer conn.Close()

	pubKey, err := os.ReadFile(m.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read the local public key file: %w", err)
	}
	pubKeyStr := strings.TrimSpace(string(pubKey))

	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create an SSH session: %w", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("grep -qxF '%s' %s || echo '%s' >> %s", pubKeyStr, remoteAuthorizedKeysPath, pubKeyStr, remoteAuthorizedKeysPath)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to add the public key to the remote authorized_keys file: %w", err)
	}

	fmt.Println("SSH public key has been successfully added to the remote server's authorized_keys file.")

	return nil
}

func (m *SSHKeyManager) generateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, m.KeyBitSize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	publicKey := &privateKey.PublicKey
	return privateKey, publicKey, nil
}

func (m *SSHKeyManager) savePrivateKey(privateKey *rsa.PrivateKey) error {
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	if err := helpers.WriteFile(m.PrivateKeyPath, privateKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write private key file: %w", err)
	}

	return nil
}

func (m *SSHKeyManager) savePublicKey(publicKey *rsa.PublicKey) error {
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to change public key from rsa to OpenSSH format: %w", err)
	}

	publicKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)
	if err := helpers.WriteFile(m.PublicKeyPath, publicKeyBytes, 0644); err != nil {
		return fmt.Errorf("failed to write public key file in OpenSSH format: %w", err)
	}

	return nil
}
