package middleware

import (
	"fmt"
	"log"
	"myapp/database"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func Authenticate2() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("Authorization")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "cookie not found",
			})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(os.Getenv("SECRET_KEY")), nil
		})
		if err != nil {
			log.Fatal(err)
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// fmt.Println(claims["foo"], claims["nbf"])

			user := database.GetUserByUsername(claims["Username"].(string))
			if *user.Username == "" {
				c.Abort()
			}

			c.Set("user", user)

			c.Next()

		} else {
			fmt.Println(err)
		}

	}
}
