package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guobinqiu/appdeployer/cmd"
	"github.com/guobinqiu/appdeployer/git"
)

type VMReq struct {
	SSHOptions     cmd.SSHOptions     `form:"ssh" json:"ssh"`
	AnsibleOptions cmd.AnsibleOptions `form:"ansible" json:"ansible"`
	DefaultOptions cmd.DefaultOptions `form:"default" json:"default"`
	GitOptions     git.GitOptions     `form:"git" json:"git"`
}

type VMDeployer struct {
}

func NewVMDeployer() *VMDeployer {
	return &VMDeployer{}
}

func (deployer *VMDeployer) Deploy(c *gin.Context) {
	var req VMReq
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": err.Error(),
		})
		return
	}

	//TODO: set default values here

	if err := cmd.VMDeploy(&req.DefaultOptions, &req.GitOptions, &req.SSHOptions, &req.AnsibleOptions); err != nil {
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
