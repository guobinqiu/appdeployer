package ssh

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net"
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
	KnownHostsPath string
	Timeout        time.Duration
}

func NewDefaultSSHKeyManager() *SSHKeyManager {
	return &SSHKeyManager{
		KeyBitSize:     2048,
		PrivateKeyPath: helpers.ExpandUser("~/.ssh/appdeployer"),
		PublicKeyPath:  helpers.ExpandUser("~/.ssh/appdeployer.pub"),
		KnownHostsPath: helpers.ExpandUser("~/.ssh/known_hosts"),
		Timeout:        0,
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

func WithKnownHostsPath(path string) func(*SSHKeyManager) {
	return func(m *SSHKeyManager) {
		m.KnownHostsPath = path
	}
}

func WithTimeout(timeout time.Duration) func(*SSHKeyManager) {
	return func(m *SSHKeyManager) {
		m.Timeout = timeout
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
	hostKey, err := m.getHostKey(m.KnownHostsPath, host)
	if err != nil {
		return fmt.Errorf("failed to get server public key from known_hosts: %w", err)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout: m.Timeout,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			if ssh.FingerprintSHA256(key) != ssh.FingerprintSHA256(hostKey) {
				return fmt.Errorf("host key does not match the expected value")
			}
			return nil
		},
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
		return fmt.Errorf("failed to add the public key to the remote server's authorized_keys file: %w", err)
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

func (m *SSHKeyManager) getHostKey(knownHostsPath string, host string) (ssh.PublicKey, error) {
	var publicKey ssh.PublicKey

	file, err := os.Open(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		line := scanner.Text()
		lineParts := strings.Split(line, " ")
		_host, encodedKey := lineParts[0], lineParts[2]

		if _host == host {
			// fmt.Println(encodedKey)

			decodedKey, err := base64.StdEncoding.DecodeString(encodedKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decode host key: %w", err)
			}

			publicKey, err = ssh.ParsePublicKey(decodedKey)
			if err != nil {
				return nil, fmt.Errorf("failed to parse SSH public key: %w", err)
			}

			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("no hosts matched")
}
