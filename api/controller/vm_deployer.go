package controller

import (
	"fmt"
	"net/http"
	"time"

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
	requestStore map[string]VMReq
}

func NewVMDeployer() *VMDeployer {
	return &VMDeployer{
		requestStore: make(map[string]VMReq),
	}
}

func (deployer *VMDeployer) Submit(c *gin.Context) {
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

	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	deployer.requestStore[requestID] = req
	c.JSON(http.StatusOK, gin.H{
		"requestID": requestID,
	})
}

func (deployer *VMDeployer) Deploy(c *gin.Context) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Streaming unsupported",
		})
		return
	}

	req, ok := deployer.requestStore[c.Query("requestID")]
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "No requestID found, call vm/submit first",
		})
		return
	}

	logCh := make(chan string)

	go func() {
		if err := cmd.VMDeploy(&req.DefaultOptions, &req.GitOptions, &req.SSHOptions, &req.AnsibleOptions, func(msg string) {
			logCh <- msg
		}); err != nil {
			logCh <- err.Error()
		}
		logCh <- "Stream ended"
		close(logCh)
	}()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	for log := range logCh {
		c.SSEvent("message", log)
		flusher.Flush()
	}
}
