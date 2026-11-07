package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Msaad668/vl-design-backend/config"
	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/Msaad668/vl-design-backend/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/oauth2/v1"
)



func SignUp(c *gin.Context) {
    var input models.SignupInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userCollection := db.GetCollection(db.Client, "users")

    var existingUser models.User
    err := userCollection.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&existingUser)
    if err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
        return
    }

    newUser := models.User{
        ID:        primitive.NewObjectID(),
        FirstName: input.FirstName,
        LastName:  input.LastName,
        Email:     input.Email,
        Password:  string(hashedPassword),
        Phone:     input.Phone,
        Address:   input.Address,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    _, err = userCollection.InsertOne(context.TODO(), newUser)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func Login(c *gin.Context) {
    var input models.LoginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userCollection := db.GetCollection(db.Client, "users")

    var user models.User
    err := userCollection.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&user)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    token, err := services.GenerateJWT(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": token})
}


func GoogleCallback(c *gin.Context) {
    code := c.Query("code")

    token, err := config.GoogleOAuthConfig.Exchange(context.TODO(), code)
    if err != nil {
        log.Printf("Failed to exchange token: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
        return
    }

    client := config.GoogleOAuthConfig.Client(context.Background(), token)

    service, err := oauth2.New(client)
    if err != nil {
        log.Printf("Failed to create OAuth2 service: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OAuth2 service"})
        return
    }

    userinfo, err := service.Userinfo.Get().Do()
    if err != nil {
        log.Printf("Failed to get user info: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
        return
    }

    userCollection := db.GetCollection(db.Client, "users")

    var existingUser models.User
    err = userCollection.FindOne(context.TODO(), bson.M{"email": userinfo.Email}).Decode(&existingUser)
    if err != nil {
        newUser := models.User{
            ID:        primitive.NewObjectID(),
            FirstName: userinfo.GivenName,
            LastName:  userinfo.FamilyName,
            Email:     userinfo.Email,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }

        _, err = userCollection.InsertOne(context.TODO(), newUser)
        if err != nil {
            log.Printf("Failed to insert new user: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
            return
        }

        jwtToken, err := services.GenerateJWT(newUser)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
            return
        }

        c.Redirect(http.StatusTemporaryRedirect, os.Getenv("GOOGLE_FRONTEND_CALLBACK_URL")+jwtToken)
        return
    }

    jwtToken, err := services.GenerateJWT(existingUser)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.Redirect(http.StatusTemporaryRedirect, os.Getenv("GOOGLE_FRONTEND_CALLBACK_URL")+jwtToken)
}