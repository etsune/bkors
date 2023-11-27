package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/etsune/bkors/server/config"
	"github.com/etsune/bkors/server/models"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	col    *mongo.Collection
	ctx    context.Context
	config config.Config
}

func NewUserService(ctx context.Context, col *mongo.Collection, config config.Config) *UserService {
	return &UserService{col, ctx, config}
}

func (s *UserService) GetUser(userId primitive.ObjectID) (models.DBUser, error) {
	var user models.DBUser
	err := s.col.FindOne(s.ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return models.DBUser{}, err
	}

	return user, nil
}

func (s *UserService) GetUserByUsername(username string) (models.DBUser, error) {
	var user models.DBUser
	err := s.col.FindOne(s.ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return models.DBUser{}, err
	}

	return user, nil
}

func (s *UserService) LoginUser(username, password string) (*http.Cookie, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v/n", user)
	if isValidPasswordHash(password, user.Password) {
		// login create cookie
		return s.CreateJwtCookie(user.Id)
	}

	return nil, err
}

func (s *UserService) RegisterUser(username, password string) (*http.Cookie, error) {
	if len(password) == 0 {
		return nil, errors.New("password can't be empty")
	}

	if len(username) > 30 {
		return nil, errors.New("username is too long")
	}

	count, err := s.col.CountDocuments(s.ctx, bson.M{"username": username})
	if err != nil {
		return nil, err
	}

	if count == 1 {
		// login user
		return s.LoginUser(username, password)
	}
	if count > 1 {
		return nil, errors.New("internal user error")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.DBUser{
		Id:       primitive.NewObjectID(),
		Username: username,
		Password: hashedPassword,
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	result, err := s.col.InsertOne(s.ctx, user)
	if err != nil {
		return nil, err
	}

	return s.CreateJwtCookie(result.InsertedID.(primitive.ObjectID))
}

func (s *UserService) CreateJwtCookie(userId primitive.ObjectID) (*http.Cookie, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid": userId,
	})

	tokenString, err := token.SignedString([]byte(s.config.AccessTokenPrivateKey))
	if err != nil {
		fmt.Println("Failed to create token")
		return nil, err
	}

	return &http.Cookie{
		Name:    "access_token",
		Value:   tokenString,
		Expires: time.Now().Add(24 * 7 * time.Hour),
	}, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func isValidPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
