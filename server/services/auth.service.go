package services

import (
	"fmt"

	"github.com/etsune/bkors/config"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService struct {
	config config.Config
}

func NewAuthService(config config.Config) *AuthService {
	return &AuthService{config}
}

func CreateToken(userId int, privateKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid": userId,
	})

	tokenString, err := token.SignedString([]byte(privateKey))
	if err != nil {
		fmt.Println("Failed to create token")
		return "", err
	}

	return tokenString, nil
}

func ParseToken(value string, privateKey string) (primitive.ObjectID, error) {
	token, err := jwt.Parse(value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(privateKey), nil
	})

	if err != nil {
		return primitive.ObjectID{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId, err := primitive.ObjectIDFromHex(claims["userid"].(string))
		return userId, err
	}

	return primitive.ObjectID{}, nil
}
