package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func MatchUserTypeToUid(c *gin.Context, userID string) (err error) {

	uid := c.GetString("user_id")

	if uid != userID {
		err = errors.New("unauthorized to access this resource")
		return err
	}

	return nil

}
