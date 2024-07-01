package ssh

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/guobinqiu/appdeployer/helpers"
	"golang.org/x/crypto/ssh"
)

type SSHKeyManager struct {
	KeyBitSize            int
	PrivateKeyPath        string
	PublicKeyPath         string
	KnownHostsPath        string
	Timeout               time.Duration
	StrictHostKeyChecking bool
}

var (
	errNoHostMatched    = errors.New("no hosts matched")
	errNoHostKeyMatched = errors.New("host key does not match the expected value")
)

func NewDefaultSSHKeyManager() *SSHKeyManager {
	return &SSHKeyManager{
		KeyBitSize:            2048,
		PrivateKeyPath:        helpers.ExpandUser("~/.ssh/appdeployer"),
		PublicKeyPath:         helpers.ExpandUser("~/.ssh/appdeployer.pub"),
		KnownHostsPath:        helpers.ExpandUser("~/.ssh/known_hosts"),
		Timeout:               0,
		StrictHostKeyChecking: true,
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
		m.PrivateKeyPath = helpers.ExpandUser(path)
	}
}

func WithPublicKeyPath(path string) func(*SSHKeyManager) {
	return func(m *SSHKeyManager) {
		m.PublicKeyPath = helpers.ExpandUser(path)
	}
}

func WithKnownHostsPath(path string) func(*SSHKeyManager) {
	return func(m *SSHKeyManager) {
		m.KnownHostsPath = helpers.ExpandUser(path)
	}
}

func WithTimeout(timeout time.Duration) func(*SSHKeyManager) {
	return func(m *SSHKeyManager) {
		m.Timeout = timeout
	}
}

func WithStrictHostKeyChecking(strictHostKeyChecking bool) func(*SSHKeyManager) {
	return func(m *SSHKeyManager) {
		m.StrictHostKeyChecking = strictHostKeyChecking
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
		Timeout: m.Timeout,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			if !m.StrictHostKeyChecking {
				return nil
			}
			hostKey, err := m.getHostKey(m.KnownHostsPath, hostname)
			if err != nil {
				if os.IsNotExist(err) || err == errNoHostMatched {
					confirmed := m.promptUserToConfirmFingerprint(hostname, key)
					if confirmed {
						return m.saveKnownHosts(hostname, key)
					}
					return err
				}
				return err
			}
			if ssh.FingerprintSHA256(key) != ssh.FingerprintSHA256(hostKey) {
				return errNoHostKeyMatched
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

func (m *SSHKeyManager) saveKnownHosts(hostname string, key ssh.PublicKey) error {
	data := []byte(fmt.Sprintf("%s %s\n", hostname, ssh.MarshalAuthorizedKey(key)))
	if err := helpers.AppendFile(m.KnownHostsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write public key file in OpenSSH format: %w", err)
	}

	fmt.Printf("SSH server's public key has been successfully added to the SSH client's %s file.\n", m.KnownHostsPath)
	return nil
}

func (m *SSHKeyManager) getHostKey(knownHostsPath string, host string) (ssh.PublicKey, error) {
	var publicKey ssh.PublicKey

	file, err := os.Open(knownHostsPath)
	if err != nil {
		fmt.Println("aaaa")
		return nil, err
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

	return nil, errNoHostMatched
}

func (m *SSHKeyManager) promptUserToConfirmFingerprint(host string, pubKey ssh.PublicKey) bool {
	fingerprint := fmt.Sprintf("%x", sha256.Sum256(pubKey.Marshal()))

	fmt.Printf("The authenticity of host '%s' can't be established.\n", host)
	fmt.Printf("RSA key fingerprint is %s.\n", fingerprint)
	fmt.Print("Are you sure you want to continue connecting (yes/no)? ")

	var userInput string
	_, err := fmt.Scanln(&userInput)
	if err != nil {
		userInput = "no"
	}

	userInput = strings.TrimSpace(strings.ToLower(userInput))
	return userInput == "yes" || userInput == "y"
}
