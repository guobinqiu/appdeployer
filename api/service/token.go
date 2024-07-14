package service

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/guobinqiu/appdeployer/api/model"
	"golang.org/x/crypto/bcrypt"
)

const (
	shortDuration = 1 * 24 * time.Hour
	longDuration  = 7 * 24 * time.Hour
)

var signingKey = []byte("孤舟蓑笠翁，独钓寒江雪")

type ClientClaims struct {
	User   *model.User
	Client *model.Client
	jwt.StandardClaims
}

func CreateJWTToken(user *model.User, client *model.Client, d time.Duration) (string, error) {
	claims := ClientClaims{
		user,
		client,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(d).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

func ParseJWTToken(tokenString string) (*model.User, *model.Client, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ClientClaims{}, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, nil, err
	}
	return token.Claims.(*ClientClaims).User, token.Claims.(*ClientClaims).Client, err
}

func UserToToken(user *model.User, client *model.Client) (accessToken string, refreshToken string) {
	accessToken, _ = CreateJWTToken(user, client, shortDuration)
	refreshToken, _ = CreateJWTToken(user, client, longDuration)
	return
}

func TokenToUser(token string) (*model.User, *model.Client, error) {
	return ParseJWTToken(token)
}

func HashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}
