package helper

import (
	"context"
	"fmt"
	"log"
	"myapp/database"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email    string
	Username string
	User_id  string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(recEmail string, recUsername string, recID string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:    recEmail,
		Username: recUsername,
		User_id:  recID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(1)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(2)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, User_id string) {

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var UpdateObj primitive.D
	UpdateObj = append(UpdateObj, bson.E{Key: "token", Value: signedToken})
	UpdateObj = append(UpdateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	upsert := true
	filter := bson.M{"user_id": User_id}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{{Key: "$set", Value: UpdateObj}},
		&opt,
	)
	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}

}

func ValidateToken(userToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(userToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	},
	)

	if err != nil {
		msg = err.Error()
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token has expired")
		msg = err.Error()
		return
	}

	return claims, msg

}
