package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddThread(c *gin.Context) {
    fmt.Println("AddThread endpoint hit")

    projectID := c.Param("projectID")
    fmt.Println("Received projectID:", projectID)

    projectObjectID, err := primitive.ObjectIDFromHex(projectID)
    if err != nil {
        fmt.Println("Error converting projectID to ObjectID:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
        return
    }

	
    role, exists := c.Get("role")
    if !exists {
		fmt.Println("No role found in context")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user role in token"})
        return
    }
	
    if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		fmt.Println("Error parsing multipart form:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
        return
    }

    message := c.PostForm("message")

    if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
        return
    }
	
    fmt.Println("Before inserting thread to DB ...")
	
	userID, exists := c.Get("userID")
	if !exists {
		fmt.Println("No userID found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid userID format"})
		return
	}
	
	userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
		return
	}
	
	thread := models.Thread{
		ID:         primitive.NewObjectID(),
		ProjectID:  projectObjectID,
		AuthorID:   userObjectID,
		Role:       role.(string),
		Message:    message,
		Documents:  []models.Document{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

    if c.Query("draft") == "true" && role != "customer" {
        thread.IsDraft = true
    }

	// handling file uploads
	files := c.Request.MultipartForm.File["documents"]
    fmt.Println("Received files:", files)
	tempDir := "./temp"
	if len(files) > 0 {
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			os.MkdirAll(tempDir, os.ModePerm)
		}

		for _, fileHeader := range files {
			filePath := filepath.Join(tempDir, fileHeader.Filename)

			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
				return
			}
			defer file.Close()

			out, err := os.Create(filePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
				return
			}
			defer out.Close()

			if _, err := io.Copy(out, file); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
				return
			}

			document := models.Document{
				Name: fileHeader.Filename,
				Size: fmt.Sprintf("%d KB", fileHeader.Size/1024),
				URL:  "/temp/" + fileHeader.Filename,
                UploadedBy: userObjectID,

			}
			thread.Documents = append(thread.Documents, document)
		}
	}

	collection := db.GetCollection(db.Client, "threads")
	if _, err := collection.InsertOne(context.TODO(), thread); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save thread"})
		return
	}

    projectCollection := db.GetCollection(db.Client, "projects")
    var project models.Project
    if err := projectCollection.FindOne(context.TODO(), bson.M{"_id": projectObjectID}).Decode(&project); err != nil {
        fmt.Println("Project not found (for notifications):", err)
    } else {
        senderID := userObjectID            
        senderRole := role.(string)

        var targetUserID primitive.ObjectID
        var subject, msgBody, recipientEmail string

        if senderRole == "designer" {
            targetUserID = project.CustomerID 
            subject = "New Thread Created by Designer"
            msgBody = fmt.Sprintf("Hello,\nA new thread was created by the designer on your project.\nMessage: %s", message)
        } else if senderRole == "customer" {
            if project.DesignerID != nil {
                targetUserID = *project.DesignerID
            } else {
                fmt.Println("No designer assigned to project, skipping notification.")
            }
            subject = "New Thread Created by Customer"
            msgBody = fmt.Sprintf("Hello,\nA new thread was created by the customer.\nMessage: %s", message)
        } else if senderRole == "admin" {
            targetUserID = project.CustomerID 
            subject = "New Thread Created by Admin"
            msgBody = fmt.Sprintf("Hello,\nAn admin created a new thread.\nMessage: %s", message)
        }


        if targetUserID != primitive.NilObjectID {
            title := "New Thread"
            if err := CreateNotification(targetUserID, senderID, title, msgBody,  &thread.ID, "thread"); err != nil {
                fmt.Println("Failed to create notification:", err)
            }
            usersCollection := db.GetCollection(db.Client, "users")
            var targetUser models.User
            err := usersCollection.FindOne(context.TODO(), bson.M{"_id": targetUserID}).Decode(&targetUser)
            if err == nil {
                recipientEmail = targetUser.Email
                if recipientEmail != "" {
                    go func() {
                        if err := SendEmail(recipientEmail, subject, msgBody); err != nil {
                            fmt.Println("Error sending email:", err)
                        }
                    }()
                }
            }
        }
    }

	c.JSON(http.StatusOK, gin.H{"message": "Thread created successfully", "thread": thread})
}

func GetThreadsForProject(c *gin.Context) {
    projectID := c.Param("projectID")

    projectObjectID, err := primitive.ObjectIDFromHex(projectID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
        return
    }

    threadsCollection := db.GetCollection(db.Client, "threads")
    usersCollection := db.GetCollection(db.Client, "users")

    role, exists := c.Get("role")
    if !exists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No role for this user"})
        return
    }

    filter := bson.M{"project_id": projectObjectID}
    
    if c.Query("draft") == "true" && role == "customer" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No role for this user"})
        return
    } else if c.Query("draft") == "true" {
        filter["is_draft"] = true
    }

    cursor, err := threadsCollection.Find(context.TODO(), filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch threads"})
        return
    }
    defer cursor.Close(context.TODO())

    var threads []models.Thread
    if err := cursor.All(context.TODO(), &threads); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse threads"})
        return
    }

    responseThreads := make([]gin.H, 0, len(threads))

    for _, thread := range threads {
        var author models.User
        err = usersCollection.FindOne(context.TODO(), bson.M{"_id": thread.AuthorID}).Decode(&author)
        if err != nil {
            author = models.User{
                FirstName: "Unknown",
                LastName:  "User",
                Email:     "",
                Role:      "Unknown",
            }
        }

        threadDocs := make([]gin.H, 0, len(thread.Documents))
        for _, doc := range thread.Documents {
            threadDocs = append(threadDocs, gin.H{
                "id":   doc.ID.Hex(),
                "name": doc.Name,
                "size": doc.Size,
                "url":  doc.URL,
            })
        }

        authorInfo := gin.H{
            "id":         author.ID.Hex(),
            "first_name": author.FirstName,
            "last_name":  author.LastName,
            "email":      author.Email,
            "role":       author.Role,
        }

        lastTwoReplies := []gin.H{}
        totalReplies := len(thread.Replies)
        if totalReplies > 0 {
            startIdx := totalReplies - 2
            if startIdx < 0 {
                startIdx = 0
            }

            recentReplies := thread.Replies[startIdx:]
            for _, r := range recentReplies {
                replyDocs := []gin.H{}
                for _, doc := range r.Documents {
                    replyDocs = append(replyDocs, gin.H{
                        "id":   doc.ID.Hex(),
                        "name": doc.Name,
                        "size": doc.Size,
                        "url":  doc.URL,
                    })
                }

                lastTwoReplies = append(lastTwoReplies, gin.H{
                    "id":       r.ID.Hex(),
                    "author":   r.AuthorName,
                    "role":     r.Role,
                    "time":     r.CreatedAt.Format("3:04 PM"),
                    "message":  r.Message,
                    "documents": replyDocs,
                })
            }
        }

        responseThreads = append(responseThreads, gin.H{
            "id":        thread.ID.Hex(),
            "time":      thread.CreatedAt.Format("3:04 PM"),
            "date":      thread.CreatedAt.Format("02 January 2006"),
            "message":   thread.Message,
            "documents": threadDocs,
            "author":    authorInfo,
            "replies":   lastTwoReplies,
        })
    }

    c.JSON(http.StatusOK, responseThreads)
}


func DeleteThread(c *gin.Context) {
    threadID := c.Param("threadID")

    threadObjectID, err := primitive.ObjectIDFromHex(threadID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
        return
    }

	
    collection := db.GetCollection(db.Client, "threads")
	thread := models.Thread{}
	err = collection.FindOne(context.TODO(), bson.M{"_id": threadObjectID}).Decode(&thread)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	currentUserID, _ := c.Get("userID")
	currentUserRole, _ := c.Get("role")
	currentUserIDStr := currentUserID.(string)
	if thread.AuthorID.Hex() != currentUserIDStr && currentUserRole != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this thread"})
		return
	}

    result, err := collection.DeleteOne(context.TODO(), bson.M{"_id": threadObjectID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete thread"})
        return
    }

    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
        return
    }

    repliesCollection := db.GetCollection(db.Client, "replies")
    _, err = repliesCollection.DeleteMany(context.TODO(), bson.M{"thread_id": threadObjectID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete associated replies"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Thread and its replies deleted successfully"})
}

func AddReplyToThread(c *gin.Context) {
    threadID := c.Param("threadID")

    threadObjectID, err := primitive.ObjectIDFromHex(threadID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
        return
    }

    userIDVal, userExists := c.Get("userID")
    if !userExists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    roleVal, roleExists := c.Get("role")
    if !roleExists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user role in token"})
        return
    }

    userIDStr, ok := userIDVal.(string)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid userID format"})
        return
    }

    userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
        return
    }

    roleStr, _ := roleVal.(string)

    if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
        return
    }

    message := c.PostForm("message")
    if message == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
        return
    }

    authorNameVal, nameExists := c.Get("name")
    authorNameStr := "Unknown"
    if nameExists {
        if nameStr, ok := authorNameVal.(string); ok {
            authorNameStr = nameStr
        }
    }

    files := c.Request.MultipartForm.File["documents"]
    tempDir := "./temp" 
    documents := []models.Document{}

    if len(files) > 0 {
        if _, err := os.Stat(tempDir); os.IsNotExist(err) {
            os.MkdirAll(tempDir, os.ModePerm)
        }

        for _, fileHeader := range files {
            filePath := filepath.Join(tempDir, fileHeader.Filename)

            file, err := fileHeader.Open()
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
                return
            }
            defer file.Close()

            out, err := os.Create(filePath)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
                return
            }
            defer out.Close()

            if _, err := io.Copy(out, file); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
                return
            }

            doc := models.Document{
                ID:   primitive.NewObjectID(),
                Name: fileHeader.Filename,
                Size: fmt.Sprintf("%d KB", fileHeader.Size/1024),
                URL:  "/temp/" + fileHeader.Filename,
            }
            documents = append(documents, doc)
        }
    }

    reply := models.Reply{
        ID:         primitive.NewObjectID(),
        ThreadID:   threadObjectID,
        AuthorID:   userObjectID,
        AuthorName: authorNameStr,
        Role:       roleStr,
        Message:    message,
        Documents:  documents,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }

    repliesCollection := db.GetCollection(db.Client, "replies")

    if _, err := repliesCollection.InsertOne(context.TODO(), reply); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save reply"})
        return
    }

    threadsCollection := db.GetCollection(db.Client, "threads")
    _, err = threadsCollection.UpdateOne(
        context.TODO(),
        bson.M{"_id": threadObjectID},
        bson.M{"$push": bson.M{"replies": reply}},
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update thread with new reply"})
        return
    }


    var thread models.Thread
    if err := threadsCollection.FindOne(context.TODO(), bson.M{"_id": threadObjectID}).Decode(&thread); err != nil {
        fmt.Println("Could not find thread to get project ID:", err)
    } else {
        projectID := thread.ProjectID

        projectCollection := db.GetCollection(db.Client, "projects")
        var project models.Project
        if err := projectCollection.FindOne(context.TODO(), bson.M{"_id": projectID}).Decode(&project); err != nil {
            fmt.Println("Project not found (for notifications):", err)
        } else {
            senderID := userObjectID
            senderRole := roleStr

            var targetUserID primitive.ObjectID
            var subject, msgBody, recipientEmail string
            if thread.IsDraft && senderRole == "customer" {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Customer isn't allowed to add a reply in the thread"})
                return
            }

            if senderRole == "designer" {
                targetUserID = project.CustomerID
                subject = "New Reply from Designer"
                msgBody = fmt.Sprintf("Hello,\nYou have a new reply on your project thread.\nMessage: %s", message)
            } else if senderRole == "customer" {
                if project.DesignerID != nil {
                    targetUserID = *project.DesignerID
                }
                subject = "New Reply from Customer"
                msgBody = fmt.Sprintf("Hello,\nThe customer has replied to the project thread.\nMessage: %s", message)
            } else if senderRole == "admin" {
                targetUserID = project.CustomerID 
                subject = "New Reply from Admin"
                msgBody = fmt.Sprintf("Hello,\nAn admin has replied to the thread.\nMessage: %s", message)
            }

            if targetUserID != primitive.NilObjectID {
                if err := CreateNotification(targetUserID, senderID, subject, msgBody, &threadObjectID, "thread"); err != nil {
                    fmt.Println("Failed to create notification:", err)
                }

                usersCollection := db.GetCollection(db.Client, "users")
                var targetUser models.User
                err := usersCollection.FindOne(context.TODO(), bson.M{"_id": targetUserID}).Decode(&targetUser)
                if err == nil && targetUser.Email != "" {
                    recipientEmail = targetUser.Email
                    go func() {
                        if err := SendEmail(recipientEmail, subject, msgBody); err != nil {
                            fmt.Println("Error sending email:", err)
                        }
                    }()
                }
            }
        }
    }


    c.JSON(http.StatusOK, gin.H{"message": "Reply added successfully", "reply": reply})
}

func GetRepliesForThread(c *gin.Context) {
    threadID := c.Param("threadID")

    threadObjectID, err := primitive.ObjectIDFromHex(threadID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
        return
    }

    collection := db.GetCollection(db.Client, "replies")

    cursor, err := collection.Find(context.TODO(), bson.M{"thread_id": threadObjectID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch replies"})
        return
    }
    defer cursor.Close(context.TODO())

    var replies []models.Reply
    if err := cursor.All(context.TODO(), &replies); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse replies"})
        return
    }

    responseReplies := make([]gin.H, 0, len(replies))
    for _, reply := range replies {
        replyDocs := make([]gin.H, 0, len(reply.Documents))
        for _, doc := range reply.Documents {
            replyDocs = append(replyDocs, gin.H{
                "id":   doc.ID.Hex(),
                "name": doc.Name,
                "size": doc.Size,
                "url":  doc.URL,
            })
        }

        responseReplies = append(responseReplies, gin.H{
            "id":        reply.ID.Hex(),
            "author":    reply.AuthorName,
            "role":      reply.Role,
            "time":      reply.CreatedAt.Format("3:04 PM"),
            "message":   reply.Message,
            "documents": replyDocs,
        })
    }

    c.JSON(http.StatusOK, responseReplies)
}

func DeleteReplyFromThread(c *gin.Context) {
    threadID := c.Param("threadID")
    replyID := c.Param("replyID")

    threadObjectID, err := primitive.ObjectIDFromHex(threadID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
        return
    }

    replyObjectID, err := primitive.ObjectIDFromHex(replyID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reply ID"})
        return
    }

    repliesCollection := db.GetCollection(db.Client, "replies")
    var reply models.Reply
    if err := repliesCollection.FindOne(context.TODO(), bson.M{"_id": replyObjectID, "thread_id": threadObjectID}).Decode(&reply); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Reply not found"})
        return
    }

    currentUserID, _ := c.Get("userID")
    currentUserRole, _ := c.Get("role")
    currentUserIDStr := currentUserID.(string)

    // here we can check some permissions if we can delete this reply
    if reply.AuthorID.Hex() != currentUserIDStr && currentUserRole != "Admin" {
        c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this reply"})
        return
    }

    result, err := repliesCollection.DeleteOne(context.TODO(), bson.M{"_id": replyObjectID, "thread_id": threadObjectID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete reply"})
        return
    }

    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Reply not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Reply deleted successfully"})
}

func GetThreadByID(c *gin.Context) {
    threadID := c.Param("threadID")

    threadObjectID, err := primitive.ObjectIDFromHex(threadID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
        return
    }

    threadsCollection := db.GetCollection(db.Client, "threads")
    usersCollection := db.GetCollection(db.Client, "users")

    var thread models.Thread
    err = threadsCollection.FindOne(context.TODO(), bson.M{"_id": threadObjectID}).Decode(&thread)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
        return
    }

    var author models.User
    err = usersCollection.FindOne(context.TODO(), bson.M{"_id": thread.AuthorID}).Decode(&author)
    if err != nil {
        author = models.User{
            FirstName: "Unknown",
            LastName:  "User",
            Email:     "",
            Role:      "Unknown",
        }
    }

    var threadDocs []gin.H
    for _, doc := range thread.Documents {
        threadDocs = append(threadDocs, gin.H{
            "id":   doc.ID.Hex(),
            "name": doc.Name,
            "size": doc.Size,
            "url":  doc.URL,
        })
    }

    authorInfo := gin.H{
        "id":         author.ID.Hex(),
        "first_name": author.FirstName,
        "last_name":  author.LastName,
        "email":      author.Email,
        "role":       author.Role,
    }


    var replies []gin.H
    for _, r := range thread.Replies {
        replyDocs := []gin.H{}
        for _, doc := range r.Documents {
            replyDocs = append(replyDocs, gin.H{
                "id":   doc.ID.Hex(),
                "name": doc.Name,
                "size": doc.Size,
                "url":  doc.URL,
            })
        }

        replies = append(replies, gin.H{
            "id":        r.ID.Hex(),
            "author":    r.AuthorName,
            "role":      r.Role,
            "time":      r.CreatedAt.Format("3:04 PM"),
            "message":   r.Message,
            "documents": replyDocs,
        })
    }

    response := gin.H{
        "id":        thread.ID.Hex(),
        "message":   thread.Message,
        "time":      thread.CreatedAt.Format("3:04 PM"),
        "date":      thread.CreatedAt.Format("02 January 2006"),
        "documents": threadDocs,
        "author":    authorInfo,
        "replies":   replies,
    }

    c.JSON(http.StatusOK, response)
}