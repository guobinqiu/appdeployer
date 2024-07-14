package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guobinqiu/appdeployer/api/service"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqToken := c.GetHeader("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"err-code": "403",
				"err-msg":  "Bearer token not in proper format",
				"path":     c.Request.URL.Path,
			})
			c.Abort()
			return
		}

		user, client, err := service.TokenToUser(strings.TrimSpace(splitToken[1]))
		if err != nil {
			log.Printf("Parsing jwt token err: %s\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"err-code": "403",
				"err-msg":  "Invalid token",
				"path":     c.Request.URL.Path,
			})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("client", client)
		c.Next()
	}
}
