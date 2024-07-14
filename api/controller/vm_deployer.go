package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guobinqiu/appdeployer/cmd"
)

type VMReq struct {
	SSHOptions     cmd.SSHOptions     `form:"ssh" json:"ssh"`
	AnsibleOptions cmd.AnsibleOptions `form:"ansible" json:"ansible"`
}

type VMDeployer struct {
}

func NewVMDeployer() *VMDeployer {
	return &VMDeployer{}
}

func (deployer *VMDeployer) Deploy(c *gin.Context) {
	var vmReq VMReq
	if err := c.ShouldBind(vmReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"msg":  "success",
		"data": vmReq,
	})
}
