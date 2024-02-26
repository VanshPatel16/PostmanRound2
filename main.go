package main

import (
	"context"
	"fmt"
	"log"
	"myapp/routes"
	"os"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

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

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	// app := controllers.NewApplication(database.UserData(database.Client, "Users"), database.PostData(database.Client, "Posts"))
	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	log.Fatal(router.Run(":" + port))

}
