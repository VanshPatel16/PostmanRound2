package middleware

import (
	"myapp/helper"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {

		clientToken := c.Request.Header.Get("token")
		// clientToken := c.GetHeader("token")

		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "No authorization header provided",
			})
			c.Abort()
			return
		}

		claims, msg := helper.ValidateToken(clientToken)
		if msg != "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			c.Abort()
			return

		}

		c.Set("email", claims.Email)
		c.Set("User_id", claims.User_id)

		c.Next()

	}
}
