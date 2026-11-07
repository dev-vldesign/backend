package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func LoadEnv() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found, relying on environment variables")
    }
}

func GetMongoURI() string {
    return os.Getenv("MONGODB_URI")
}

func GetJWTSecret() string {
    return os.Getenv("JWT_SECRET")
}

var GoogleOAuthConfig *oauth2.Config

func SetupGoogleOAuth() {
	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_API_CLIENT_ID"),     
		ClientSecret: os.Getenv("GOOGLE_API_CLIENT_SECRET"), 
		RedirectURL:  os.Getenv("GOOGLE_API_REDIRECT_URL"), 
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}