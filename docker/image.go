package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/guobinqiu/deployer/helpers"
)

const DOCKERHUB = "https://index.docker.io/v1/"

type DockerOptions struct {
	AppDir       string
	Dockerfile   string
	Dockerconfig string
	Registry     string
	Username     string
	Password     string
	Repository   string
	Tag          string
}

func (opts DockerOptions) Image() (string, error) {
	if helpers.IsBlank(opts.Registry) {
		return "", errors.New("docker.registry is required")
	}
	if helpers.IsBlank(opts.Repository) {
		return "", errors.New("docker.repository is required")
	}

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

	return builder.String(), nil
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

func (ds *DockerService) BuildImage(ctx context.Context, opts DockerOptions) error {
	// 创建一个上下文（context）包含Dockerfile和其他构建所需文件
	buildCtx, err := archive.TarWithOptions(opts.AppDir, &archive.TarOptions{})
	if err != nil {
		panic(err)
	}
	defer buildCtx.Close()

	img, err := opts.Image()
	if err != nil {
		return err
	}

	// 构建镜像的选项
	buildOptions := types.ImageBuildOptions{
		Dockerfile: opts.Dockerfile,
		Context:    buildCtx,
		Tags:       []string{img},
	}

	resp, err := ds.cli.ImageBuild(ctx, buildCtx, buildOptions)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取并处理构建过程中的输出流，这里通常会打印到控制台
	io.Copy(os.Stdout, resp.Body)
	return nil
}

func (ds *DockerService) PushImage(ctx context.Context, opts DockerOptions) error {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	defer cli.Close()

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

	img, err := opts.Image()
	if err != nil {
		return err
	}

	// 推送镜像
	pushResp, err := ds.cli.ImagePush(ctx, img, image.PushOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}
	defer pushResp.Close()

	// 打印推送日志
	body, err := io.ReadAll(pushResp)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}
