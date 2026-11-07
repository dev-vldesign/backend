package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Ventor represents a vendor/contractor in the system
type Ventor struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VentorName string             `bson:"ventor_name" json:"ventor_name" binding:"required"`
	Phone      string             `bson:"phone" json:"phone" binding:"required"`
	Email      string             `bson:"email" json:"email" binding:"required,email"`
	Website    string             `bson:"website" json:"website"`
	City       string             `bson:"city" json:"city" binding:"required"`
	State      string             `bson:"state" json:"state" binding:"required"`
	Country    string             `bson:"country" json:"country" binding:"required"`
	Role       string             `bson:"role" json:"role" binding:"required"` // "contractor" or "supplier"

	// File uploads (optional)
	Logo    string `bson:"logo,omitempty" json:"logo,omitempty"`
	License string `bson:"license,omitempty" json:"license,omitempty"`

	// Contractor-specific fields
	YearsOfExperience string   `bson:"years_of_experience,omitempty" json:"years_of_experience,omitempty"`
	CompanySize       string   `bson:"company_size,omitempty" json:"company_size,omitempty"`
	ExpectedBudget    string   `bson:"expected_budget,omitempty" json:"expected_budget,omitempty"`
	VentorServices    []string `bson:"ventor_services,omitempty" json:"ventor_services,omitempty"`

	// Common fields
	LicenseNumber     string    `bson:"license_number" json:"license_number" binding:"required"`
	LicenseExpiration string    `bson:"license_expiration" json:"license_expiration" binding:"required"`
	VentorType        string    `bson:"ventor_type" json:"ventor_type" binding:"required"`
	HowDidYouHear     string    `bson:"how_did_you_hear" json:"how_did_you_hear"`
	CreatedAt         time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time `bson:"updated_at" json:"updated_at"`
}

// CreateVentorInput represents the input for creating a new ventor
type CreateVentorInput struct {
	VentorName string `form:"ventor_name" binding:"required"`
	Phone      string `form:"phone" binding:"required"`
	Email      string `form:"email" binding:"required,email"`
	Website    string `form:"website"`
	City       string `form:"city" binding:"required"`
	State      string `form:"state" binding:"required"`
	Country    string `form:"country" binding:"required"`
	Role       string `form:"role" binding:"required"` // "contractor" or "supplier"

	// Contractor-specific fields
	YearsOfExperience string   `form:"years_of_experience,omitempty"`
	CompanySize       string   `form:"company_size,omitempty"`
	ExpectedBudget    string   `form:"expected_budget,omitempty"`
	VentorServices    []string `form:"ventor_services,omitempty"`

	// Common fields
	LicenseNumber     string `form:"license_number" binding:"required"`
	LicenseExpiration string `form:"license_expiration" binding:"required"`
	VentorType        string `form:"ventor_type" binding:"required"`
	HowDidYouHear     string `form:"how_did_you_hear"`
}
