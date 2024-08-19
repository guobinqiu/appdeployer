package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guobinqiu/appdeployer/cmd"
	"github.com/guobinqiu/appdeployer/docker"
	"github.com/guobinqiu/appdeployer/git"
)

type KubeReq struct {
	DockerOptions  docker.DockerOptions `form:"docker" json:"docker"`
	KubeOptions    cmd.KubeOptions      `form:"kube" json:"kube"`
	GitOptions     git.GitOptions       `form:"git" json:"git"`
	DefaultOptions cmd.DefaultOptions   `form:"default" json:"default"`
}

type KubeDeployer struct {
}

func NewKubeDeployer() *KubeDeployer {
	return &KubeDeployer{}
}

func (deployer *KubeDeployer) Deploy(c *gin.Context) {
	var req KubeReq
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": err.Error(),
		})
		return
	}

	if err := cmd.KubeDeploy(&req.DefaultOptions, &req.GitOptions, &req.KubeOptions, &req.DockerOptions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"msg":  "success",
		"data": req,
	})
}
