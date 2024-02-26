package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	Username      *string            `json:"username" validate:"required"`
	First_name    *string            `json:"first_name"`
	Last_name     *string            `json:"last_name"`
	Email         *string            `json:"email" validate:"required"`
	Password      *string            `json:"password" validate:"required"`
	Token         *string            `json:"token"`
	Refresh_token *string            `json:"refresh_token"`
	User_id       string             `json:"user_id"`
	Posts         []string           `json:"posts"`
	Following     []string           `json:"following"`
	Followers     []string           `json:"followers"`
	Avatar        string             `json:"avatar"`
	Bookmark      []string           `json:"bookmark"`
}

type UserProfile struct {
	Username   *string  `json:"username"`
	Email      *string  `json:"email"`
	User_id    string   `json:"user_id"`
	First_name *string  `json:"first_name"`
	Last_name  *string  `json:"last_name"`
	Posts      []string `json:"posts"`
	Following  []string `json:"following"`
	Followers  []string `json:"followers"`
	Avatar     string   `json:"avatar"`
	Bookmark   []string `json:"bookmark"`
}

type GoogleUser struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	Hd            string `json:"hd"`
	Locale        string `json:"locale"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"`
}

type Post struct {
	ID        primitive.ObjectID
	Post_id   string
	Username  *string
	Pictures  []string
	Caption   *string
	Comments  []Comment
	LikedBy   []string
	LikeCount uint64
	Tags      []string
}

type Comment struct {
	Username string `bson:"username"`
	Body     string `bson:"body"`
}
