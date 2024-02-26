package database

import (
	"context"
	"fmt"
	"myapp/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func DeletePost(postId string, email string) error {
	postCollection := OpenCollection(Client, "posts")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	//find post,and delete it from db
	_, err := postCollection.DeleteOne(ctx, bson.M{"post_id": postId})

	if err != nil {

		return err
	}

	//if post has been deleted from posts collection,we also remove it from user's collection

	userCollection := OpenCollection(Client, "users")

	var user models.User

	_ = userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	i := -1
	for j := 0; j < len(user.Posts); j++ {

		if user.Posts[j] == postId {
			i = j
			break
		}

	}

	if i < 0 {

		err := fmt.Errorf("this post is not made by the user")
		return err
	}
	if len(user.Posts) > 1 && i < len(user.Posts)-1 {

		user.Posts = append(user.Posts[:i], user.Posts[i+1]) //deleted the postid from posts of the user
	} else if i == len(user.Posts)-1 && len(user.Posts) > 1 {
		user.Posts = user.Posts[:i]
	} else {

		var empty []string
		user.Posts = empty

	}

	//now update the user

	if UpdateUser(email, bson.M{"posts": user.Posts}) != nil {
		err := fmt.Errorf("could not update user")
		return err
	}

	return nil
}

func GetPostById(postID string) (post models.Post) {
	postCollection := OpenCollection(Client, "posts")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	_ = postCollection.FindOne(ctx, bson.M{"post_id": postID}).Decode(&post)

	return post

}
