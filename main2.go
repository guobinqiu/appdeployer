package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	// 设置SSH客户端配置
	config := &ssh.ClientConfig{
		User: "guobin",
		Auth: []ssh.AuthMethod{
			ssh.Password("111111"),
		},
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 创建SSH客户端
	client, err := ssh.Dial("tcp", "192.168.1.9:22", config)
	if err != nil {
		panic(err)
	}

	// 关闭SSH客户端
	defer client.Close()

	// 读取控制机的公钥
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	publicKeyBytes, err := os.ReadFile(filepath.Join(usr.HomeDir, ".ssh", "id_rsa.pub"))
	if err != nil {
		panic(err)
	}

	publicKey := strings.TrimSpace(string(publicKeyBytes))

	session, err := client.NewSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// 执行shell命令将公钥追加到authorized_keys文件
	authorizedKeysPath := ".ssh/authorized_keys"
	cmd := fmt.Sprintf("echo '%s' >> %s", publicKey, authorizedKeysPath)
	if err := session.Run(cmd); err != nil {
		panic(err)
	}

	fmt.Println("SSH公钥已成功添加到受控机的authorized_keys文件中")
}
