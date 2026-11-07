package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/db/pipelines"
	"github.com/Msaad668/vl-design-backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDesignerDashboard(c *gin.Context) {
	// Extract and validate user ID
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

	// Get MongoDB collection
	threadsCollection := db.GetCollection(db.Client, "projects")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build the aggregation pipeline
	pipeline := pipelines.GetDashboardData(userObjectID)

	// Execute the aggregation
	cursor, err := threadsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard data"})
		return
	}
	defer cursor.Close(ctx)

	// Process results
	var results []struct {
		StatusSummary    []map[string]interface{} `bson:"status_summary"`
		PerformanceGraph []map[string]interface{} `bson:"performance_graph"`
		RecentProjects   []map[string]interface{} `bson:"recent_projects"`
		CurrentTasks     []map[string]interface{} `bson:"current_tasks"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode dashboard data"})
		return
	}

	// Handle case when designer has no projects
	if len(results) == 0 || len(results[0].StatusSummary) == 0 {
		// Return empty dashboard with zero values
		response := gin.H{
			"status_summary": gin.H{
				"overdue":              0,
				"in_progress":          0,
				"projects_in_progress": 0,
				"workload":             "0%",
			},
			"performance": gin.H{
				"period_days": 360,
				"data":        []interface{}{},
			},
			"recent_projects": []interface{}{},
			"current_tasks":   []interface{}{},
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Format the response
	response := gin.H{
		"status_summary": gin.H{
			"overdue":            results[0].StatusSummary[0]["overdue"],
			"in_progress":        results[0].StatusSummary[0]["in_progress"],
			"completed_projects": results[0].StatusSummary[0]["total_projects"],
			"workload":           results[0].StatusSummary[0]["workload"],
		},
		"performance": gin.H{
			"period_days": 360,
			"data":        utils.FormatPerformanceData(results[0].PerformanceGraph),
		},
		"recent_projects": results[0].RecentProjects,
		"current_tasks":   results[0].CurrentTasks,
	}

	c.JSON(http.StatusOK, response)
}
