package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RespondWithJSON is a helper function to send JSON responses.
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, `{"message": "Error marshalling JSON response"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondWithError is a helper for sending consistent JSON error responses.
func RespondWithError(c *gin.Context, code int, message string, err error) {
	log.Printf("Error: %s, Details: %v", message, err) // Log the error for debugging
	c.JSON(code, gin.H{"message": message})
}
