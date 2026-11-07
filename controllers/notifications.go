package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateNotification(
    userID, 
    senderID primitive.ObjectID, 
    title, 
    message string, 
    contextID *primitive.ObjectID, 
    contextType string,
) error {

    notificationsCollection := db.GetCollection(db.Client, "notifications")

    notification := models.Notification{
        ID:          primitive.NewObjectID(),
        UserID:      userID,
        SenderID:    senderID,
        Title:       title,
        Message:     message,
        ContextID:   contextID,
        ContextType: contextType,
        IsSeen:      false,
        CreatedAt:   time.Now(),
    }

    _, err := notificationsCollection.InsertOne(context.TODO(), notification)
    if err != nil {
        log.Println("Error inserting notification:", err)
        return err
    }
    return nil
}

func SendEmail(to, subject, body string) error {
    from := os.Getenv("SMTP_EMAIL")
    password := os.Getenv("SMTP_PASSWORD")
    smtpHost := "smtp.gmail.com"
    smtpPort := "587"

    msg := "From: " + from + "\n" +
        "To: " + to + "\n" +
        "Subject: " + subject + "\n\n" +
        body

    auth := smtp.PlainAuth("", from, password, smtpHost)

    err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))
    if err != nil {
        fmt.Println("Failed to send email:", err)
        return err
    }

    fmt.Println("Email sent successfully to", to)
    return nil
}

func GetNotificationsForUser(c *gin.Context) {
    userIDVal, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    userIDStr, ok := userIDVal.(string)
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID format"})
        return
    }

    userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
        return
    }

    notificationsCollection := db.GetCollection(db.Client, "notifications")

    findOptions := options.Find()
    findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

    cursor, err := notificationsCollection.Find(
        context.TODO(),
        bson.M{"user_id": userObjectID},
        findOptions,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
        return
    }
    defer cursor.Close(context.TODO())

    var notifications []models.Notification
    if err := cursor.All(context.TODO(), &notifications); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse notifications"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func MarkNotificationsAsSeen(c *gin.Context) {
    userIDVal, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    userIDStr, ok := userIDVal.(string)
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID format"})
        return
    }

    userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
        return
    }

    var requestBody struct {
        NotificationIDs []string `json:"notification_ids"`
    }

    if err := c.ShouldBindJSON(&requestBody); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    if len(requestBody.NotificationIDs) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No notification IDs provided"})
        return
    }

    var objectIDs []primitive.ObjectID
    for _, idStr := range requestBody.NotificationIDs {
        oid, err := primitive.ObjectIDFromHex(idStr)
        if err == nil {
            objectIDs = append(objectIDs, oid)
        }
    }
    if len(objectIDs) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No valid notification IDs"})
        return
    }

    notificationsCollection := db.GetCollection(db.Client, "notifications")

    filter := bson.M{
        "_id":     bson.M{"$in": objectIDs},
        "user_id": userObjectID,
    }
    update := bson.M{
        "$set": bson.M{
            "is_seen": true,
        },
    }

    result, err := notificationsCollection.UpdateMany(context.TODO(), filter, update)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notifications"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message":          "Notifications marked as seen",
        "modified_count":   result.ModifiedCount,
        "matched_count":    result.MatchedCount,
        "acknowledged":     result.MatchedCount > 0,
        "notification_ids": requestBody.NotificationIDs,
    })
}
