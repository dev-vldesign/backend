package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Msaad668/vl-design-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ValidateVentorInput validates the ventor input data
func ValidateVentorInput(input *models.CreateVentorInput) error {
	// Validate role
	if input.Role != "contractor" && input.Role != "supplier" {
		return fmt.Errorf("role must be either 'contractor' or 'supplier'")
	}

	// Validate contractor-specific fields
	if input.Role == "contractor" {
		if input.YearsOfExperience == "" {
			return fmt.Errorf("years of experience is required for contractors")
		}
		if input.CompanySize == "" {
			return fmt.Errorf("company size is required for contractors")
		}
		if input.ExpectedBudget == "" {
			return fmt.Errorf("expected budget is required for contractors")
		}
		if len(input.VentorServices) == 0 {
			return fmt.Errorf("ventor services are required for contractors")
		}
	}

	return nil
}

// ValidateAndSaveLogo handles logo file upload validation and saving
func ValidateAndSaveLogo(c *gin.Context) (string, error) {
	logoFile, logoHandler, err := c.Request.FormFile("logo")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil // Logo is optional
		}
		return "", fmt.Errorf("error processing logo upload: %v", err)
	}
	defer logoFile.Close()

	// Validate image file type
	allowedImageTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	contentType := logoHandler.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return "", fmt.Errorf("logo must be an image file (JPEG, PNG, GIF, WEBP), received: %s", contentType)
	}

	// Generate unique filename for logo
	logoExt := filepath.Ext(logoHandler.Filename)
	logoFilename := fmt.Sprintf("logo_%s%s", primitive.NewObjectID().Hex(), logoExt)
	logoFilePath := filepath.Join("uploads", "ventors", "logos", logoFilename)

	// Ensure the uploads/ventors/logos directory exists
	err = os.MkdirAll(filepath.Dir(logoFilePath), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create logo directory: %v", err)
	}

	// Create and save logo file
	out, err := os.Create(logoFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to save logo: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, logoFile)
	if err != nil {
		return "", fmt.Errorf("failed to write logo content: %v", err)
	}

	// Normalize the path for database storage
	logoPath := "/" + filepath.ToSlash(logoFilePath)
	return logoPath, nil
}

// ValidateAndSaveLicense handles license file upload validation and saving
func ValidateAndSaveLicense(c *gin.Context) (string, error) {
	licenseFile, licenseHandler, err := c.Request.FormFile("license")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil // License is optional
		}
		return "", fmt.Errorf("error processing license upload: %v", err)
	}
	defer licenseFile.Close()

	// Validate PDF file type
	contentType := licenseHandler.Header.Get("Content-Type")
	if contentType != "application/pdf" {
		return "", fmt.Errorf("license must be a PDF file, received: %s", contentType)
	}

	// Generate unique filename for license
	licenseExt := filepath.Ext(licenseHandler.Filename)
	licenseFilename := fmt.Sprintf("license_%s%s", primitive.NewObjectID().Hex(), licenseExt)
	licenseFilePath := filepath.Join("uploads", "ventors", "licenses", licenseFilename)

	// Ensure the uploads/ventors/licenses directory exists
	err = os.MkdirAll(filepath.Dir(licenseFilePath), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create license directory: %v", err)
	}

	// Create and save license file
	out, err := os.Create(licenseFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to save license: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, licenseFile)
	if err != nil {
		return "", fmt.Errorf("failed to write license content: %v", err)
	}

	// Normalize the path for database storage
	licensePath := "/" + filepath.ToSlash(licenseFilePath)
	return licensePath, nil
}

// HandleFileUploads processes both logo and license file uploads
func HandleFileUploads(c *gin.Context) (string, string, error) {
	// Handle logo upload
	logoPath, err := ValidateAndSaveLogo(c)
	if err != nil {
		return "", "", err
	}

	// Handle license upload
	licensePath, err := ValidateAndSaveLicense(c)
	if err != nil {
		return "", "", err
	}

	return logoPath, licensePath, nil
}

// CreateVentorFromInput creates a Ventor struct from CreateVentorInput
func CreateVentorFromInput(input *models.CreateVentorInput, logoPath, licensePath string) models.Ventor {
	ventor := models.Ventor{
		ID:                primitive.NewObjectID(),
		VentorName:        input.VentorName,
		Phone:             input.Phone,
		Email:             input.Email,
		Website:           input.Website,
		City:              input.City,
		State:             input.State,
		Country:           input.Country,
		Role:              input.Role,
		Logo:              logoPath,
		License:           licensePath,
		LicenseNumber:     input.LicenseNumber,
		LicenseExpiration: input.LicenseExpiration,
		VentorType:        input.VentorType,
		HowDidYouHear:     input.HowDidYouHear,
	}

	// Add contractor-specific fields only if role is contractor
	if input.Role == "contractor" {
		ventor.YearsOfExperience = input.YearsOfExperience
		ventor.CompanySize = input.CompanySize
		ventor.ExpectedBudget = input.ExpectedBudget
		ventor.VentorServices = input.VentorServices
	}

	return ventor
}
