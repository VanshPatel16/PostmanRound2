package controllers

import (
	"context"
	"myapp/database"
	"myapp/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func FollowAnotherUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		//make sure user has logged in
		user1, _ := c.Get("user")
		userID := user1.(models.User).User_id

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
		defer cancel()
		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if *user.Token == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "user has logged out,please login to follow other users",
			})
			return
		}

		//we will obtain the user username to be followed,User's own username

		var req struct {
			UsernameToFollow string
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "only enter the username to be followed",
			})
			return
		}

		//search both the users from DB and perform the operations

		var UserToBeFollowed models.User

		userCollection := database.OpenCollection(database.Client, "users")
		ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		userCollection.FindOne(ctx, bson.M{"username": req.UsernameToFollow}).Decode(&UserToBeFollowed)
		defer cancel()

		//check if user already there
		i := -1
		for j, val := range UserToBeFollowed.Followers {
			if val == *user.Username {
				i = j
				break
			}

		}

		if i == -1 {
			//means user is not a follower
			//so we add him as a follower
			UserToBeFollowed.Followers = append(UserToBeFollowed.Followers, *user.Username)
			//now update UserToBeFollowed in database
			_, err := userCollection.UpdateOne(ctx, bson.M{"username": UserToBeFollowed.Username}, bson.M{"$set": UserToBeFollowed})
			defer cancel()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "could not update the user to be followed",
				})
				return
			}

			//now update user as well
			user.Following = append(user.Following, *UserToBeFollowed.Username)
			_, err = userCollection.UpdateOne(ctx, bson.M{"username": user.Username}, bson.M{"$set": user})
			defer cancel()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "could not update the user who is following",
				})
				return
			}

		} else {
			//means user is a follower
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "You already follow this user",
			})
			return
		}

		//if all is good,return a msg that he has followed

		c.JSON(http.StatusOK, gin.H{
			"msg":            "succesfully followed the user",
			"following list": user.Following,
		})

	}
}
