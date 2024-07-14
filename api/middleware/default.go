package middleware

import (
	"github.com/gin-gonic/gin"
)

func Default() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
		// 处理请求
		c.Next()
	}
}
