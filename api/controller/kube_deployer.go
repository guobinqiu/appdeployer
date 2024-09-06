package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guobinqiu/appdeployer/cmd"
	"github.com/guobinqiu/appdeployer/docker"
	"github.com/guobinqiu/appdeployer/git"
	"github.com/guobinqiu/appdeployer/helpers"
	"github.com/guobinqiu/appdeployer/kube"
)

type KubeReq struct {
	DockerOptions  docker.DockerOptions `form:"docker" json:"docker"`
	KubeOptions    cmd.KubeOptions      `form:"kube" json:"kube"`
	GitOptions     git.GitOptions       `form:"git" json:"git"`
	DefaultOptions cmd.DefaultOptions   `form:"default" json:"default"`
}

type KubeDeployer struct {
	logCh        chan string
	requestStore map[string]KubeReq
}

func NewKubeDeployer() *KubeDeployer {
	return &KubeDeployer{
		logCh:        make(chan string),
		requestStore: make(map[string]KubeReq),
	}
}

func (deployer *KubeDeployer) Submit(c *gin.Context) {
	var req KubeReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": err.Error(),
		})
		return
	}

	helpers.SetDefault(&req.DockerOptions.Dockerconfig, "~/.docker/config.json")
	helpers.SetDefault(&req.DockerOptions.Dockerfile, "./Dockerfile")
	helpers.SetDefault(&req.DockerOptions.Registry, docker.DOCKERHUB)
	helpers.SetDefault(&req.DockerOptions.Tag, "latest")
	helpers.SetDefault(&req.KubeOptions.Kubeconfig, "~/.kube/config")
	helpers.SetDefault(&req.KubeOptions.IngressOptions.TLS, false)
	helpers.SetDefault(&req.KubeOptions.IngressOptions.SelfSigned, false)
	helpers.SetDefault(&req.KubeOptions.IngressOptions.SelfSignedYears, 1)
	helpers.SetDefault(&req.KubeOptions.ServiceOptions.Port, int32(8000))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.Replicas, int32(1))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.Port, int32(8000))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.RollingUpdate.MaxSurge, "1")
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.RollingUpdate.MaxUnavailable, "0")
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.LivenessProbe.Enabled, false)
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.LivenessProbe.Type, kube.ProbeTypeHTTPGet)
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.LivenessProbe.Path, "/")
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.LivenessProbe.Scheme, "http")
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.LivenessProbe.InitialDelaySeconds, int32(0))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.LivenessProbe.TimeoutSeconds, int32(1))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.LivenessProbe.PeriodSeconds, int32(10))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.LivenessProbe.SuccessThreshold, int32(1))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.LivenessProbe.FailureThreshold, int32(3))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.ReadinessProbe.Enabled, false)
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.ReadinessProbe.Type, kube.ProbeTypeHTTPGet)
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.ReadinessProbe.Path, "/")
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.ReadinessProbe.Scheme, "http")
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.ReadinessProbe.InitialDelaySeconds, int32(0))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.ReadinessProbe.TimeoutSeconds, int32(1))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.ReadinessProbe.PeriodSeconds, int32(10))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.ReadinessProbe.SuccessThreshold, int32(1))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.ReadinessProbe.FailureThreshold, int32(3))
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.VolumeMount.Enabled, false)
	helpers.SetDefault(&req.KubeOptions.DeploymentOptions.VolumeMount.MountPath, "/app/data")
	helpers.SetDefault(&req.KubeOptions.HpaOptions.Enabled, false)
	helpers.SetDefault(&req.KubeOptions.HpaOptions.MinReplicas, int32(1))
	helpers.SetDefault(&req.KubeOptions.HpaOptions.MaxReplicas, int32(10))
	helpers.SetDefault(&req.KubeOptions.HpaOptions.CPURate, int32(50))
	helpers.SetDefault(&req.KubeOptions.PvcOptions.AccessMode, "readwriteonce")
	helpers.SetDefault(&req.KubeOptions.PvcOptions.StorageClassName, "openebs-hostpath")
	helpers.SetDefault(&req.KubeOptions.PvcOptions.StorageSize, "1G")

	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	deployer.requestStore[requestID] = req
	c.JSON(http.StatusOK, gin.H{
		"requestID": requestID,
	})
}

func (deployer *KubeDeployer) Deploy(c *gin.Context) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Streaming unsupported",
		})
		return
	}

	requestID := c.Query("requestID")
	req, ok := deployer.requestStore[requestID]
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "No requestID found, call kube/submit first",
		})
		return
	}

	go func() {
		if err := cmd.KubeDeploy(&req.DefaultOptions, &req.GitOptions, &req.KubeOptions, &req.DockerOptions, func(msg string) {
			deployer.logCh <- msg
		}); err != nil {
			deployer.logCh <- err.Error()
		}
	}()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	for log := range deployer.logCh {
		c.SSEvent("message", log)
		flusher.Flush()
	}

	deployer.logCh <- "Stream ended"
	close(deployer.logCh)
	delete(deployer.requestStore, requestID)
}
