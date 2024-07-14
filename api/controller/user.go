package controller

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guobinqiu/appdeployer/api/model"
	"github.com/guobinqiu/appdeployer/api/service"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const DefaultGrantType = "password"

type LoginReq struct {
	ClientKey    string `form:"client_id" json:"client_id" binding:"required"`
	ClientSecret string `form:"client_secret" json:"client_secret" binding:"required"`
	Username     string `form:"username" json:"username" binding:"required"`
	Password     string `form:"password" json:"password" binding:"required"`
	GrantType    string `form:"grant_type" json:"grant_type" binding:"required"`
}

type UserController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{
		db: db,
	}
}

func (ctl *UserController) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err-code": "400",
			"err-msg":  err.Error(),
		})
		return
	}

	if strings.ToLower(req.GrantType) != DefaultGrantType {
		c.JSON(http.StatusBadRequest, gin.H{
			"err-code": "400",
			"err-msg":  "Only password mode is supported",
		})
		return
	}

	client, err := ctl.GetLoginClient(req.ClientKey, req.ClientSecret)
	if err != nil {
		log.Printf("Get login client err: %s\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"err-code": "400",
			"err-msg":  "Login failed",
		})
		return
	}

	user, err := ctl.GetLoginUser(req.Username, req.Password, client.ID)
	if err != nil {
		log.Printf("Get login user err: %s\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"err-code": "400",
			"err-msg":  "Login failed",
		})
		return
	}

	accessToken, _ := service.UserToToken(user, client)
	c.JSON(http.StatusCreated, gin.H{
		"err-code": "0",
		"err-msg":  "success",
		"payload": gin.H{
			"token": accessToken,
		},
	})
}

func (ctl *UserController) GetLoginClient(clientKey, clientSecret string) (*model.Client, error) {
	var client model.Client
	var loginClientErr = errors.New("invalid client key or secret")

	err := ctl.db.First(&client, "client_key = ?", clientKey).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, loginClientErr
	}

	if !verify(client.ClientSecret, clientSecret) {
		return &client, loginClientErr
	}

	return &client, nil
}

func (ctl *UserController) GetLoginUser(username, password string, clientID int) (*model.User, error) {
	var user model.User
	var loginUserErr = errors.New("invalid username or password")

	err := ctl.db.First(&user, "username = ? and client_id = ?", username, clientID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, loginUserErr
	}

	if !verify(user.Password, password) {
		return &user, loginUserErr
	}

	return &user, nil
}

func verify(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
