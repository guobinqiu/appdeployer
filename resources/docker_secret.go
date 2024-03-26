package resources

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/guobinqiu/deployer/docker"
	"github.com/guobinqiu/deployer/helpers"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DockerSecretOptions struct {
	ApplicationName string
	Namespace       string
	docker.DockerOptions
}

func CreateOrUpdateDockerSecret(clientset *kubernetes.Clientset, ctx context.Context, opts DockerSecretOptions) error {
	dockerconfigjson, err := buildDockerAuthConfig(opts.DockerOptions)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "docker-" + opts.ApplicationName,
			Namespace: opts.Namespace,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": dockerconfigjson,
		},
	}

	if _, err = clientset.CoreV1().Secrets(opts.Namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		log.Println("docker secret resource successfully updated")
	} else {
		log.Println("docker secret resource successfully created")
	}

	return nil
}

func buildDockerAuthConfig(opts docker.DockerOptions) ([]byte, error) {
	var dockerConfig map[string]interface{}

	if !helpers.IsBlank(opts.Registry) && !helpers.IsBlank(opts.Username) && !helpers.IsBlank(opts.Password) {
		// 构造Docker配置信息
		dockerConfig = map[string]interface{}{
			"auths": map[string]interface{}{
				opts.Registry: map[string]string{
					"auth": getAuthString(opts.Username, opts.Password),
				},
			},
		}
	} else if !helpers.IsBlank(opts.Dockerconfig) {
		//读取Docker配置文件
		configData, err := os.ReadFile(opts.Dockerconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to read Docker config file: %w", err)
		}

		if err := json.Unmarshal(configData, &dockerConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Docker config JSON: %w", err)
		}

		// 确保配置中有对应Registry的auth数据
		registryData, ok := dockerConfig["auths"].(map[string]interface{})
		if !ok || registryData[opts.Registry] != nil {
			return nil, fmt.Errorf("no auth data found for registry %s in the provided Docker config", opts.Registry)
		}
	} else {
		return nil, fmt.Errorf("neither username/password nor config file specified")
	}

	// 将Docker配置JSON对象转换为[]byte
	dockerConfigJSON, err := json.Marshal(dockerConfig)
	if err != nil {
		return nil, err
	}
	return dockerConfigJSON, nil
}

func getAuthString(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}
