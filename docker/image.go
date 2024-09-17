package docker

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/guobinqiu/appdeployer/helpers"
)

const DOCKERHUB = "https://index.docker.io/v1/"

type DockerOptions struct {
	AppDir       string `form:"appdir" json:"appdir"`
	Dockerfile   string `form:"dockerfile" json:"dockerfile"`
	Dockerconfig string `form:"dockerconfig" json:"dockerconfig"`
	Registry     string `form:"registry" json:"registry"`
	Username     string `form:"username" json:"username"`
	Password     string `form:"password" json:"password"`
	Repository   string `form:"repository" json:"repository"`
	Tag          string `form:"tag" json:"tag"`
}

func (opts DockerOptions) Validate() error {
	if helpers.IsBlank(opts.Registry) {
		return errors.New("docker.registry is required")
	}
	if helpers.IsBlank(opts.Repository) {
		return errors.New("docker.repository is required")
	}
	return nil
}

func (opts DockerOptions) Image() string {
	builder := new(strings.Builder)
	if opts.Registry != DOCKERHUB {
		builder.WriteString(opts.Registry)
		builder.WriteByte('/')
	}
	builder.WriteString(opts.Repository)
	if !helpers.IsBlank(opts.Tag) {
		builder.WriteByte(':')
		builder.WriteString(opts.Tag)
	}
	return builder.String()
}

type DockerService struct {
	cli *client.Client
}

// 创建一个新的Docker客户端实例
func NewDockerService() (*DockerService, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	return &DockerService{cli: cli}, nil
}

func (ds *DockerService) Close() error {
	return ds.cli.Close()
}

type BuildMessage struct {
	Stream string `json:"stream"`
	Status string `json:"status"`
	Aux    struct {
		ID string `json:"ID"`
	} `json:"aux"`
}

func (ds *DockerService) BuildImage(ctx context.Context, opts DockerOptions, logHandler func(msg string)) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	// 创建一个上下文（context）包含Dockerfile和其他构建所需文件
	buildCtx, err := archive.TarWithOptions(opts.AppDir, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("unable to create tar archive from directory '%s': %w", opts.AppDir, err)
	}
	defer buildCtx.Close()

	// 构建镜像的选项
	buildOptions := types.ImageBuildOptions{
		Dockerfile: opts.Dockerfile,
		Context:    buildCtx,
		Tags:       []string{opts.Image()},
	}

	resp, err := ds.cli.ImageBuild(ctx, buildCtx, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to build Docker image: %v", err)
	}
	defer resp.Body.Close()

	// 逐行打印响应流
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		var msg BuildMessage
		json.Unmarshal([]byte(line), &msg)
		logHandler(msg.Stream)
	}

	return nil
}

type PushMessage struct {
	Status         string                 `json:"status"`
	ProgressDetail map[string]interface{} `json:"progressDetail"`
	ID             string                 `json:"id"`
	Aux            map[string]interface{} `json:"aux"`
}

func (ds *DockerService) PushImage(ctx context.Context, opts DockerOptions, logHandler func(msg string)) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	// 登录到Docker registry
	authConfig := registry.AuthConfig{
		Username:      opts.Username,
		Password:      opts.Password,
		ServerAddress: opts.Registry,
	}

	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	// 推送镜像
	resp, err := ds.cli.ImagePush(ctx, opts.Image(), image.PushOptions{RegistryAuth: authStr})
	if err != nil {
		return fmt.Errorf("failed to marshal auth configuration to JSON: %v", err)
	}
	defer resp.Close()

	// 逐行打印响应流
	scanner := bufio.NewScanner(resp)
	for scanner.Scan() {
		line := scanner.Text()
		var msg PushMessage
		json.Unmarshal([]byte(line), &msg)
		logHandler(msg.Status)
	}

	return nil
}
