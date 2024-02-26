package database

import (
	"context"
	"fmt"
	"myapp/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func DeleteUserAccount(email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userCollection := OpenCollection(Client, "users")
	var user models.User

	_ = userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if *user.Token == "" {
		err := fmt.Errorf("login to delete your account")
		return err
	}

	_, err := userCollection.DeleteOne(ctx, bson.M{"email": email})
	if err != nil {
		return err
	}

	return nil

}

func UpdateUser(email string, UpdateData bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userCollection := OpenCollection(Client, "users")
	filter := bson.M{"email": email}
	update := bson.M{"$set": UpdateData}
	_, err := userCollection.UpdateOne(ctx, filter, update)

	if err != nil {
		return err
	}

	return nil
}

func CheckUserInDB(email string) bool {

	userCollection := OpenCollection(Client, "users")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var user models.User

	_ = userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if *user.Email == "" {
		return false
	}

	return true
}

func GetUserByUsername(username string) (user models.User) {

	userCollection := OpenCollection(Client, "users")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	_ = userCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)

	return user

}

func GetUserByEmail(email string) (user models.User) {

	userCollection := OpenCollection(Client, "users")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	_ = userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	return user

}
