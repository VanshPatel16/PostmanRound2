package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"myapp/database"
	"myapp/helper"
	"myapp/models"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")
var validate = validator.New()

func NewOAuth2Config() (*oauth2.Config, error) {
	providerURL := fmt.Sprintf("https://" + os.Getenv("GOOGLE_CLIENT_DOMAIN") + "/")
	provider, err := oidc.NewProvider(context.Background(), providerURL)
	if err != nil {
		return nil, fmt.Errorf("could not create new provider")
	}

	var Oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		Scopes:       []string{"profile,email,photo"},
		Endpoint:     provider.Endpoint(),
	}

	return Oauth2Config, nil

}

func GenerateRandomString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	s := base64.StdEncoding.EncodeToString(b)

	return s, nil
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)

	if err != nil {
		log.Panic(err)
	}

	return string(bytes)

}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "email or password is incorrect"
		check = false

	}

	return check, msg

}

func CheckIfLoggedIn(email string) bool {

	userCollection := database.OpenCollection(database.Client, "users")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	//get user from collection
	var user models.User
	_ = userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	return *user.Token != ""
}

func GetOwnUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		// userID := c.MustGet("User_id")

		// var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		// var user models.User

		// err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
		// defer cancel()
		// if err != nil {

		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"error": err.Error(),
		// 	})
		// 	return
		// }

		user, _ := c.Get("user")

		if *user.(models.User).Token == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "user has logged out,please login to view profile",
			})
			return
		}

		var myProfile models.UserProfile

		myProfile.Email = user.(models.User).Email
		myProfile.Username = user.(models.User).Username
		myProfile.First_name = user.(models.User).First_name
		myProfile.Last_name = user.(models.User).Last_name
		myProfile.User_id = user.(models.User).User_id
		myProfile.Avatar = user.(models.User).Avatar
		myProfile.Followers = user.(models.User).Followers
		myProfile.Following = user.(models.User).Following
		myProfile.Posts = user.(models.User).Posts
		myProfile.Bookmark = user.(models.User).Bookmark

		c.JSON(http.StatusOK, myProfile)
		// c.HTML(200, "/static/userpage.html", nil)

	}
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		defer cancel()

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": validationErr.Error(),
			})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "error occurred while checking for the user",
			})
			return
		}
		password := HashPassword(*user.Password)
		user.Password = &password

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "An account with this email alredy exists",
			})
			return
		}
		var posts []string
		var followers []string
		var following []string

		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.Username, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken
		user.Posts = posts
		user.Followers = followers
		user.Following = following
		(user.Avatar) = "https://i.stack.imgur.com/l60Hf.png" //default user avatar

		resultInsertionNumber, err := userCollection.InsertOne(ctx, user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Username is taken",
			})
			return

		}

		c.JSON(http.StatusOK, gin.H{
			"msg":          "User created succesfully",
			"insertion_id": resultInsertionNumber,
		})
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		var foundUser models.User
		if err := c.BindJSON(&user); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{

				"error": err.Error(),
			})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "email or password is incorrect",
			})
			return
		}

		passIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)

		if !passIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "user not found",
			})
			return

		}
		token, refreshtoken, _ := helper.GenerateAllTokens(*user.Email, *foundUser.Username, foundUser.User_id)

		helper.UpdateAllTokens(token, refreshtoken, foundUser.User_id)

		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("Authorization", token, int(time.Hour.Seconds()), "/", "", false, true)

		c.JSON(http.StatusOK, foundUser)
		// // c.JSON(http.StatusOK, gin.H{
		// // 	"token": foundUser.Token,
		// // })

	}
}

func DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		// email := c.MustGet("email")
		email := user.(models.User).Email

		err := database.DeleteUserAccount(*email)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not delete user",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg": "Account deleted succesfully",
		})

	}
}

func UpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// email := c.MustGet("email")
		user, _ := c.Get("user")
		email := user.(models.User).Email

		loginStatus := CheckIfLoggedIn(*email)

		if !loginStatus {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "must be logged in",
			})
			return
		}

		var UpdateData bson.M

		if err := c.BindJSON(&UpdateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request body",
			})
			return
		}

		err := database.UpdateUser(*email, UpdateData)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not update data",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg": "user updated succesfully",
		})

	}
}

func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {

		// email := c.MustGet("email")
		user, _ := c.Get("user")
		email := user.(models.User).Email
		loginStatus := CheckIfLoggedIn(*email)

		if !loginStatus {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "must be logged in to logout",
			})
			return

		}

		updateData := bson.M{"token": ""}

		err := database.UpdateUser(*email, updateData)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not log user out",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg": "User logged out succesfully",
		})

	}
}

func GoogleLoginInitiate() gin.HandlerFunc {

	return func(c *gin.Context) {
		GoogleOAuthConfig, _ := NewOAuth2Config()
		state, _ := GenerateRandomString()
		url := GoogleOAuthConfig.AuthCodeURL(state)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

func GoogleLoginCallback() gin.HandlerFunc {
	return func(c *gin.Context) {

		code := c.Query("code")

		GoogleOAuthConfig, _ := NewOAuth2Config()

		token, err := GoogleOAuthConfig.Exchange(c, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "could not exchange token")
			return
		}
		if !token.Valid() {
			c.JSON(http.StatusBadRequest, "token expired")
			return
		}

		client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
		resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not create client",
			})
			return
		}
		defer resp.Body.Close()
		var userInfo models.GoogleUser
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not get user",
			})
			return
		}

		fmt.Println(userInfo)

		c.Redirect(http.StatusTemporaryRedirect, "/profile")
	}
}

// func GoogleLoginCallback() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		code := c.Query("code")
// 		token, err := GoogleOAuthConfig.Exchange(context.Background(), code)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"error": "failed to exchange token",
// 			})

// 			return
// 		}
// 		if !token.Valid() {
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"error": "token expired",
// 			})
// 			return
// 		}
// 		client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
// 		resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"error": "could not create client",
// 			})
// 			return
// 		}
// 		defer resp.Body.Close()
// 		var userInfo models.GoogleUser
// 		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"error": "could not get user",
// 			})
// 			return
// 		}
// 		//we have user details now

// 		//check if user email is in db

// 		c.JSON(http.StatusOK, gin.H{
// 			"msg": "nice done",
// 		})

// 		result := database.CheckUserInDB(userInfo.Email)

// 		if result {

// 			//if email is there then we give him new token and give the message the user is logged in
// 			var userObtained models.User
// 			userCollection := database.OpenCollection(database.Client, "users")
// 			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
// 			userCollection.FindOne(ctx, bson.M{"email": userInfo.Email}).Decode(&userObtained)
// 			defer cancel()
// 			token, refreshtoken, _ := helper.GenerateAllTokens(*userObtained.Email, *userObtained.Username, userObtained.User_id)
// 			helper.UpdateAllTokens(token, refreshtoken, userObtained.User_id)

// 			c.JSON(http.StatusOK, gin.H{
// 				"msg":          "logged in succesfully",
// 				"user details": userObtained,
// 			})

// 			return

// 		} else {

// 			// if email is not there then we create a new account for user to sign in
// 			var user models.User
// 			var posts []string
// 			var followers []string
// 			var following []string
// 			user.First_name = &userInfo.GivenName
// 			user.Last_name = &userInfo.FamilyName
// 			user.Username = &userInfo.Name
// 			user.ID = primitive.NewObjectID()
// 			user.User_id = user.ID.Hex()
// 			token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.Username, user.User_id)
// 			user.Token = &token
// 			user.Refresh_token = &refreshToken
// 			user.Posts = posts
// 			user.Followers = followers
// 			user.Following = following
// 			(user.Avatar) = userInfo.Picture //default user avatar

// 			var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
// 			resultInsertionNumber, err := userCollection.InsertOne(ctx, user)
// 			defer cancel()
// 			if err != nil {
// 				c.JSON(http.StatusBadRequest, gin.H{
// 					"error": "Username is taken",
// 				})

// 				c.JSON(http.StatusOK, gin.H{
// 					"msg":          "User created succesfully",
// 					"insertion_id": resultInsertionNumber,
// 				})

// 				return
// 			}

// 		}
// 	}
// }

func UploadAvatar() gin.HandlerFunc {
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
		var AvatarUrl bson.M

		if err := c.BindJSON(&AvatarUrl); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "enter only the avatar url",
			})
			return
		}
		// email := c.MustGet("email")
		err := database.UpdateUser(*email, AvatarUrl)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg": "profile picture updated succesfully",
		})

	}
}
