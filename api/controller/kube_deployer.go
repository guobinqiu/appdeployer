package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guobinqiu/appdeployer/cmd"
	"github.com/guobinqiu/appdeployer/docker"
)

type KubeReq struct {
	DockerOptions docker.DockerOptions `form:"docker" json:"docker"`
	KubeOptions   cmd.KubeOptions      `form:"kube" json:"kube"`
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

	c.JSON(http.StatusCreated, gin.H{
		"msg":  "success",
		"data": req,
	})
}
