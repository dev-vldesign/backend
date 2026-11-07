package services

import (
	"time"

	"github.com/Msaad668/vl-design-backend/config"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
    UserID string `json:"user_id"` 
    Email  string `json:"email"`
    Role   string `json:"role"`
    Name   string `json:"first_name"`
    jwt.StandardClaims
}

func GenerateJWT(user models.User) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        UserID:    user.ID.Hex(),
        Email:     user.Email,
        Role:      user.Role,
        Name:      user.FirstName,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    jwtKey := GetJWTKey()

    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        return "", err
    }

    return tokenString, nil
}

func GetJWTKey() []byte {
    return []byte(config.GetJWTSecret())
}