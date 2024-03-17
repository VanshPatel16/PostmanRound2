package controllers

import (
	"context"
	"fmt"
	"myapp/database"
	"myapp/models"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func MakeAPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		email := user.(models.User).Email
		if !CheckIfLoggedIn(*email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "must be logged in to post",
			})
			return

		} //checked if the user is logged in.
		//if he is logged in,create a new post and take inputs from user

		if c.Request.ParseMultipartForm(10<<20) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not parse form",
			})
		}
		form, _ := c.MultipartForm()

		images := form.File["postimg"]
		if images == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "no images found",
			})
			return
		}

		var newpost models.Post
		// if err := c.ShouldBindJSON(&newpost); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"error": "incorrect format",
		// 	})
		// 	return
		// }

		newpost.Caption = c.PostForm("caption")
		newpost.Tags = c.PostFormArray("tags")

		//input captured

		//save the image
		for _, file := range images {

			filePath := filepath.Join("images", file.Filename)
			newpost.Path = append(newpost.Path, filePath)
			err := c.SaveUploadedFile(file, filePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": err,
				})
				return
			}

		}
		//now fill other details of the post
		var comments []models.Comment
		var likedby []string
		newpost.ID = primitive.NewObjectID()
		newpost.Post_id = newpost.ID.Hex()
		newpost.LikeCount = 0
		newpost.Comments = comments
		newpost.LikedBy = likedby

		//now we require the user details

		var poster models.User

		userCollection := database.OpenCollection(database.Client, "users")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		userCollection.FindOne(ctx, bson.M{"email": *email}).Decode(&poster)
		defer cancel()
		//now we update the user in db
		newpost.Username = poster.Username
		poster.Posts = append(poster.Posts, newpost.Post_id)
		updatedData := bson.M{
			"posts": poster.Posts,
		}
		err := database.UpdateUser(*poster.Email, updatedData)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "could not udate user in db",
			})
			return
		}

		//now we add the posts in the posts collection

		postCollection := database.OpenCollection(database.Client, "posts")

		_, err = postCollection.InsertOne(ctx, newpost)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not add post to the db",
			})
			return
		}

		//if all is good then

		c.JSON(http.StatusOK, gin.H{
			"msg": "post is posted",
		})

	}
}

func DeleteAPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		email := user.(models.User).Email

		if !CheckIfLoggedIn(*email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "must be logged in to post",
			})
			return

		} //checked if the user is logged in.

		//if he is logged in,take only post id as input

		var postID struct {
			Post_id string `json:"post_id"`
		}

		if err := c.BindJSON(&postID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request body",
			})
			return
		}

		var postTodelete = database.GetPostById(postID.Post_id)

		for _, val := range postTodelete.Path {
			if os.Remove(val) != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "could not delete img",
				})
			}
		}

		err := database.DeletePost(postID.Post_id, *email)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg": "post deleted succesfully",
		})

	}
}

func UpdatePost() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		email := user.(models.User).Email
		if !CheckIfLoggedIn(*email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "must be logged in to update post",
			})
			return

		} //checked if the user is logged in.

		postId := c.Param("post_id")
		var UpdatedPost bson.M

		if err := c.BindJSON(&UpdatedPost); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request body",
			})
			return
		}
		fmt.Printf("%v\n", UpdatedPost)
		fmt.Println(postId)

		postCollection := database.OpenCollection(database.Client, "posts")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		_, err := postCollection.UpdateOne(ctx, bson.M{"post_id": postId}, bson.M{"$set": UpdatedPost})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not update post",
			})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, gin.H{
			"msg": "post updated succesfully",
		})

	}
}

func ReadAPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		email := user.(models.User).Email

		if !CheckIfLoggedIn(*email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "must be logged in to view a post",
			})
			return
		}

		var postID struct {
			PostID string `json:"post_id"`
		}

		if c.BindJSON(&postID) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not bind json",
			})
			return
		}
		// fmt.Println(postID.PostID)
		// fmt.Println(email)

		postCollection := database.OpenCollection(database.Client, "posts")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var post models.Post
		_ = postCollection.FindOne(ctx, bson.M{"post_id": postID.PostID}).Decode((&post))

		c.JSON(http.StatusOK, post)
	}
}

func ReadAllPostOfAUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		email := user.(models.User).Email
		if !CheckIfLoggedIn(*email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "must be logged in to view a post",
			})
			return
		}

		var Username struct {
			Username string
		}

		if c.BindJSON(&Username) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not bind json",
			})
			return
		}

		userFound := database.GetUserByUsername(Username.Username)
		var userFeed []models.Post
		for _, val := range userFound.Posts {
			userFeed = append(userFeed, database.GetPostById(val))
		}

		c.JSON(http.StatusOK, userFeed)

	}
}

func LikeAPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		email := user.(models.User).Email

		if !CheckIfLoggedIn(*email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "must be logged in to view a post",
			})
			return
		}

		//we ask for a post id

		var PostID struct {
			PostID string `json:"post_id"`
		}

		if c.BindJSON(&PostID) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not bind json",
			})
			return
		}

		var post models.Post = database.GetPostById(PostID.PostID)

		post.LikedBy = append(post.LikedBy, *user.(models.User).Username)

		postCollection := database.OpenCollection(database.Client, "posts")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		//find post,and delete it from db

		update := bson.M{
			"$set": bson.M{"likecount": post.LikeCount + 1, "likedby": post.LikedBy},
		}
		_, err := postCollection.UpdateOne(ctx, bson.M{"post_id": PostID.PostID}, update)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not update post",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg": "post liked succesfully",
		})

	}
}

func CommentOnPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		email := user.(models.User).Email

		if !CheckIfLoggedIn(*email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "must be logged in to comment",
			})
			return
		}
		post_ID := c.Param("post_id")

		var CommentBody struct {
			Body string
		}

		if c.BindJSON(&CommentBody) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not bind json",
			})
			return
		}

		var Comment models.Comment

		Comment.Username = *user.(models.User).Username
		Comment.Body = CommentBody.Body

		postCollection := database.OpenCollection(database.Client, "posts")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		initialPost := database.GetPostById(post_ID)

		initialPost.Comments = append(initialPost.Comments, Comment)

		_, err := postCollection.UpdateOne(ctx, bson.M{"post_id": post_ID}, bson.M{"$set": bson.M{"comments": initialPost.Comments}})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not update post",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg": "Comment posted sucesfully",
		})

	}
}

func SearchPostsByTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		var Tag struct {
			Tags []string
		}

		if c.BindJSON(&Tag) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not bind json",
			})
			return
		}
		postCollection := database.OpenCollection(database.Client, "posts")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var MatchingPosts []models.Post
		filter := bson.M{"tags": bson.M{"$in": Tag.Tags}}
		cursor, err := postCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var post models.Post
			if err := cursor.Decode(&post); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			MatchingPosts = append(MatchingPosts, post)
		}

		c.JSON(http.StatusOK, gin.H{
			"posts": MatchingPosts,
		})

	}
}

func BookmarkPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		user1, _ := c.Get("user")
		email := user1.(models.User).Email

		if !CheckIfLoggedIn(*email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "must be logged in to comment",
			})
			return
		}

		var PostId struct {
			Post_id string
		}

		if c.BindJSON(&PostId) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not bind json",
			})
			return
		}

		user := database.GetUserByEmail(*email)

		user.Bookmark = append(user.Bookmark, PostId.Post_id)

		err := database.UpdateUser(*email, bson.M{"bookmark": user.Bookmark})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not add bookmark",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg": "Post bookmarked sucesfully",
		})

	}
}

func PostContent() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		email := user.(models.User).Email
		loginStatus := CheckIfLoggedIn(*email)

		if !loginStatus {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "must be logged in to update profile",
			})
			return
		}

		var postID struct {
			PostID string `json:"post_id"`
		}

		if c.BindJSON(&postID) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not bind json",
			})
			return
		}

		postCollection := database.OpenCollection(database.Client, "posts")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var post models.Post
		_ = postCollection.FindOne(ctx, bson.M{"post_id": postID.PostID}).Decode((&post))

		for i := range post.Path {
			c.File(post.Path[i])
		}

	}
}
