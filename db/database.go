package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Msaad668/vl-design-backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectDB() {
    clientOptions := options.Client().ApplyURI(config.GetMongoURI())

    client, err := mongo.NewClient(clientOptions)
    if err != nil {
        log.Fatal("Error creating MongoDB client:", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err = client.Connect(ctx)
    if err != nil {
        log.Fatal("Error connecting to MongoDB:", err)
    }

    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal("Error pinging MongoDB:", err)
    }

    fmt.Println("Connected to MongoDB!")
    Client = client
}

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
    return client.Database("vl-design").Collection(collectionName) 
}

// Create indexes when initializing your database connection
// Run this once during application initialization
func EnsureIndexes(client *mongo.Client) error {
    dbName := os.Getenv("MONGODB_NAME")
    collection := client.Database(dbName).Collection("projects")

    // List of all required indexes
    requiredIndexes := []mongo.IndexModel{
        // Designer-based indexes
        {
            Keys: bson.D{
                {Key: "designer_id", Value: 1},
                {Key: "project_start_date", Value: -1},
            },
            Options: options.Index().SetName("designer_start_date_idx"),
        },
        {
            Keys: bson.D{
                {Key: "second_designer_id", Value: 1},
                {Key: "project_start_date", Value: -1},
            },
            Options: options.Index().SetName("second_designer_start_date_idx"),
        },
        {
            Keys: bson.D{
                {Key: "designer_id", Value: 1},
                {Key: "created_at", Value: -1},
            },
            Options: options.Index().SetName("designer_recent_projects_idx"),
        },

        // Status-based indexes
        {
            Keys: bson.D{{Key: "status", Value: 1}},
            Options: options.Index().SetName("status_idx"),
        },
        {
            Keys: bson.D{
                {Key: "status", Value: 1},
                {Key: "project_end_date", Value: 1},
            },
            Options: options.Index().
                SetName("overdue_projects_idx").
                SetPartialFilterExpression(bson.M{
                    "status": bson.M{"$ne": "completed"},
                }),
        },

        // Time-based indexes
        {
            Keys: bson.D{{Key: "created_at", Value: -1}},
            Options: options.Index().SetName("created_at_desc_idx"),
        },
        {
            Keys: bson.D{{Key: "updated_at", Value: -1}},
            Options: options.Index().SetName("updated_at_desc_idx"),
        },

        // Task-related indexes
        {
            Keys: bson.D{{Key: "project_tasks.task_status", Value: 1}},
            Options: options.Index().
                SetName("active_tasks_idx").
                SetPartialFilterExpression(bson.M{
                    "project_tasks.task_status": bson.M{"$ne": "done"},
                }),
        },
        {
            Keys: bson.D{{Key: "project_tasks.task_assignee", Value: 1}},
            Options: options.Index().SetName("task_assignee_idx"),
        },
        {
            Keys: bson.D{
                {Key: "project_tasks.task_status", Value: 1},
                {Key: "project_tasks.task_assignee", Value: 1},
            },
            Options: options.Index().SetName("task_status_assignee_combo_idx"),
        },
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Validate all indexes first
    for _, index := range requiredIndexes {
        if err := validateIndex(index); err != nil {
            return fmt.Errorf("invalid index configuration: %v", err)
        }
    }

    // Get existing indexes
    cursor, err := collection.Indexes().List(ctx)
    if err != nil {
        return fmt.Errorf("failed to list indexes: %v", err)
    }

    // Check which indexes already exist
    existingIndexes := make(map[string]bool)
    for cursor.Next(ctx) {
        var index bson.M
        if err := cursor.Decode(&index); err != nil {
            return fmt.Errorf("failed to decode index: %v", err)
        }
        if name, ok := index["name"].(string); ok {
            existingIndexes[name] = true
        }
    }

    // Create missing indexes with concurrency control
    sem := make(chan struct{}, 3) // Limit concurrent index creations
    errChan := make(chan error, len(requiredIndexes))
    var wg sync.WaitGroup

    for _, index := range requiredIndexes {
        name := *index.Options.Name
        if existingIndexes[name] {
            continue
        }

        wg.Add(1)
        go func(idx mongo.IndexModel) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()

            start := time.Now()
            _, err := collection.Indexes().CreateOne(ctx, idx)
            if err != nil {
                errChan <- fmt.Errorf("failed to create index %s: %v", *idx.Options.Name, err)
                return
            }
            log.Printf("Successfully created index %s in %v", *idx.Options.Name, time.Since(start))
        }(index)
    }

    go func() {
        wg.Wait()
        close(errChan)
    }()

    // Collect any errors
    for err := range errChan {
        if err != nil {
            return err
        }
    }

    return nil
}

func validateIndex(index mongo.IndexModel) error {
    if index.Options == nil {
        return fmt.Errorf("index options cannot be nil")
    }
    if index.Options.Name == nil {
        return fmt.Errorf("index name must be specified")
    }
    if *index.Options.Name == "" {
        return fmt.Errorf("index name cannot be empty")
    }
    keys, ok := index.Keys.(bson.D)
    if !ok || len(keys) == 0 {
        return fmt.Errorf("index must have at least one key")
    }
    return nil
}