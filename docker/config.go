package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/guobinqiu/deployer/helpers"
)

func BuildDockerAuthConfig(opts DockerOptions) ([]byte, error) {
	var dockerConfig map[string]interface{}

	if !helpers.IsBlank(opts.Username) && !helpers.IsBlank(opts.Password) {
		// 构造Docker配置信息
		dockerConfig = map[string]interface{}{
			"auths": map[string]interface{}{
				opts.Registry: map[string]string{
					"auth": getAuthString(opts.Username, opts.Password),
				},
			},
		}
	} else if !helpers.IsBlank(opts.Configfile) {
		//读取Docker配置文件
		configData, err := os.ReadFile(opts.Configfile)
		if err != nil {
			return nil, fmt.Errorf("failed to read Docker config file: %w", err)
		}

		if err := json.Unmarshal(configData, &dockerConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Docker config JSON: %w", err)
		}

		// 确保配置中有对应 Registry 的 auth 数据
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
