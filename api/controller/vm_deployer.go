package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guobinqiu/appdeployer/cmd"
	"github.com/guobinqiu/appdeployer/git"
	"github.com/guobinqiu/appdeployer/helpers"
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
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": err.Error(),
		})
		return
	}

	helpers.SetDefault(&req.SSHOptions.Port, 22)
	helpers.SetDefault(&req.SSHOptions.AuthorizedKeysPath, "~/.ssh/authorized_keys")
	helpers.SetDefault(&req.SSHOptions.PrivatekeyPath, "~/.ssh/appdeployer")
	helpers.SetDefault(&req.SSHOptions.PublickeyPath, "~/.ssh/appdeployer.pub")
	helpers.SetDefault(&req.SSHOptions.KnownHostsPath, "~/.ssh/known_hosts")
	helpers.SetDefault(&req.SSHOptions.StrictHostKeyChecking, true)
	helpers.SetDefault(&req.AnsibleOptions.Hosts, "localhost")
	helpers.SetDefault(&req.AnsibleOptions.InstallDir, "~/workspace")

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
