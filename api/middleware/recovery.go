package middleware

import (
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"runtime/debug"
)

func Recovery() gin.HandlerFunc {
	return RecoveryWithWriter(gin.DefaultErrorWriter)
}

func RecoveryWithWriter(out io.Writer) gin.HandlerFunc {
	var logger *log.Logger
	if out != nil {
		logger = log.New(out, "\n\n\x1b[31m", log.LstdFlags)
	}

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if logger != nil {
					httprequest, _ := httputil.DumpRequest(c.Request, false)
					reset := string([]byte{27, 91, 48, 109})
					logger.Printf("[Recovery] panic recovered:\n\n%s%s\n\n%s%s", httprequest, err, debug.Stack(), reset)
				}
				c.JSON(http.StatusInternalServerError, gin.H{
					"statusCode": http.StatusInternalServerError,
					"message":    "internal server error",
				})
			}
		}()
		c.Next() // execute all the handlers
	}
}
