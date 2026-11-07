package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Project struct {
	ID                    primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Address               string              `bson:"address" json:"address"`
	CustomerNeedsHelpWith string              `bson:"customer_needs_help_with" json:"customer_needs_help_with"`
	ProjectArea           string              `bson:"project_area" json:"project_area"`
	PropertySize          string              `bson:"property_size" json:"property_size"`
	UniqueChallenges      []string            `bson:"unique_challenges" json:"unique_challenges"`
	YardFeatures          []string            `bson:"yard_features" json:"yard_features"`
	ProjectTimeline       string              `bson:"project_timeline" json:"project_timeline"`
	SelectedPackage       string              `bson:"selected_package" json:"selected_package"`
	Status                string              `bson:"status" json:"status"`
	IsPaid                bool                `bson:"is_paid" json:"is_paid"`
	CreatedAt             time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time           `bson:"updated_at" json:"updated_at"`
	CustomerID            primitive.ObjectID  `bson:"customer_id" json:"customer_id"`
	DesignerID            *primitive.ObjectID `bson:"designer_id,omitempty" json:"designer_id,omitempty"`
	SecondDesignerID      *primitive.ObjectID `bson:"second_designer_id,omitempty" json:"second_designer_id,omitempty"`
	OrderID               *primitive.ObjectID `bson:"order_id,omitempty" json:"order_id,omitempty"`
	ProjectDuration       string              `bson:"project_duration" json:"project_duration"`
	ProjectStartDate      time.Time           `bson:"project_start_date" json:"project_start_date"`
	ProjectEndDate        time.Time           `bson:"project_end_date" json:"project_end_date"`
	ProjectTasks          []ProjectTask       `bson:"project_tasks" json:"project_tasks"`
}

type ProjectTask struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TaskName      string             `bson:"task_name" json:"task_name"`
	TaskDuration  string             `bson:"task_duration" json:"task_duration"`
	TaskStartDate time.Time          `bson:"task_start_date" json:"task_start_date"`
	TaskEndDate   time.Time          `bson:"task_end_date" json:"task_end_date"`

	TaskAssignee *primitive.ObjectID `bson:"task_assignee,omitempty" json:"task_assignee,omitempty"`

	TaskStatus string        `bson:"task_status" json:"task_status"`
	Children   []ProjectTask `bson:"children,omitempty" json:"children,omitempty"`
}