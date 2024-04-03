package main

import (
	"os"

	"github.com/guobinqiu/deployer/ansible"
)

func main() {
	homeDir := os.Getenv("HOME")
	keyManager := ansible.NewSSHKeyManager(homeDir, "deployer", 2048)
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
