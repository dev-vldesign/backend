package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AdminUpdateProjectRequest struct {
	DesignerID       *string `json:"designer_id,omitempty"`
	SecondDesignerID *string `json:"second_designer_id,omitempty"`
	Status           *string `json:"status,omitempty"`
	IsPaid           *bool   `json:"is_paid,omitempty"`
}

func AdminUpdateProject(c *gin.Context) {
    projectIDStr := c.Param("projectID")
    projectID, err := primitive.ObjectIDFromHex(projectIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
        return
    }

    var req AdminUpdateProjectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    updateFields := bson.M{}

    if req.Status != nil && *req.Status != "" {
        updateFields["status"] = *req.Status
    }
    if req.IsPaid != nil {
        updateFields["is_paid"] = *req.IsPaid
    }

    designerIDProvided := (req.DesignerID != nil && *req.DesignerID != "")
    secondDesignerIDProvided := (req.SecondDesignerID != nil && *req.SecondDesignerID != "")

    switch {
    case designerIDProvided && secondDesignerIDProvided:
        // both are provided => set both
        designerObjID, err := primitive.ObjectIDFromHex(*req.DesignerID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid designer ID"})
            return
        }
        secondDesignerObjID, err := primitive.ObjectIDFromHex(*req.SecondDesignerID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid second designer ID"})
            return
        }
        updateFields["designer_id"] = designerObjID
        updateFields["second_designer_id"] = secondDesignerObjID

    case designerIDProvided && !secondDesignerIDProvided:
        // only primary designer => set primary, set null to second
        designerObjID, err := primitive.ObjectIDFromHex(*req.DesignerID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid designer ID"})
            return
        }
        updateFields["designer_id"] = designerObjID
        updateFields["second_designer_id"] = nil

    case secondDesignerIDProvided && !designerIDProvided:
        // only second => set second, set primary to null
        secondDesignerObjID, err := primitive.ObjectIDFromHex(*req.SecondDesignerID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid second designer ID"})
            return
        }
        updateFields["designer_id"] = nil
        updateFields["second_designer_id"] = secondDesignerObjID

    default:
        // neither is provided => do nothing
    }

    if len(updateFields) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields provided for update"})
        return
    }

    updateFields["updated_at"] = time.Now()

    collection := db.GetCollection(db.Client, "projects")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    result, err := collection.UpdateOne(ctx,
        bson.M{"_id": projectID},
        bson.M{"$set": updateFields},
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
        return
    }
    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
        return
    }

    var updatedProject models.Project
    if err := collection.FindOne(ctx, bson.M{"_id": projectID}).Decode(&updatedProject); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated project"})
        return
    }

    c.JSON(http.StatusOK, updatedProject)
}

type AdminUpdateProjectTaskRequest struct {
	TaskAssignee *string `json:"task_assignee,omitempty"`
	TaskStatus   *string `json:"task_status,omitempty"`
}

func updateTaskInSlice(tasks []models.ProjectTask, taskID primitive.ObjectID, req AdminUpdateProjectTaskRequest) (bool, []models.ProjectTask) {
	updated := false
	for i, task := range tasks {
		if task.ID == taskID {
			if req.TaskAssignee != nil {
				objID, err := primitive.ObjectIDFromHex(*req.TaskAssignee)
				if err == nil {
					tasks[i].TaskAssignee = &objID
				}
			}
			if req.TaskStatus != nil {
				tasks[i].TaskStatus = *req.TaskStatus
			}
			updated = true
			return updated, tasks
		}
		if len(task.Children) > 0 {
			var childUpdated bool
			childUpdated, tasks[i].Children = updateTaskInSlice(task.Children, taskID, req)
			if childUpdated {
				updated = true
				return updated, tasks
			}
		}
	}
	return updated, tasks
}

func AdminUpdateProjectTask(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	taskIDStr := c.Param("taskID")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	taskID, err := primitive.ObjectIDFromHex(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req AdminUpdateProjectTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := db.GetCollection(db.Client, "projects")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var project models.Project
	if err := collection.FindOne(ctx, bson.M{"_id": projectID}).Decode(&project); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	updated, updatedTasks := updateTaskInSlice(project.ProjectTasks, taskID, req)
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	project.ProjectTasks = updatedTasks
	project.UpdatedAt = time.Now()

	_, err = collection.UpdateOne(ctx, bson.M{"_id": projectID}, bson.M{"$set": bson.M{
		"project_tasks": project.ProjectTasks,
		"updated_at":    project.UpdatedAt,
	}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project task updated successfully"})
}

func GetAllUsers(c *gin.Context) {
    collection := db.GetCollection(db.Client, "users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    findOptions := options.Find().SetProjection(bson.M{"password": 0})

    filter := bson.M{"role": "designer"}

    cursor, err := collection.Find(ctx, filter, findOptions)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
        return
    }
    defer cursor.Close(ctx)

    var users []models.User
    if err := cursor.All(ctx, &users); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse users"})
        return
    }

    for i := range users {
        users[i].Password = ""
    }

    c.JSON(http.StatusOK, users)
}
