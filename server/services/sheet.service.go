package services

import (
	"context"
	"errors"
	"github.com/etsune/bkors/server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SheetService struct {
	col *mongo.Collection
	ctx context.Context
}

func NewSheetService(ctx context.Context, col *mongo.Collection) *SheetService {
	return &SheetService{col, ctx}
}

func (s *SheetService) Get(dict string, num int) (*models.DBPage, error) {
	if num < 1 {
		return nil, errors.New("num <1")
	}
	if num > 10000 {
		return nil, errors.New("num >10.000")
	}

	filter := bson.M{"dict": dict, "num": num}
	var page models.DBPage
	if err := s.col.FindOne(s.ctx, filter).Decode(&page); err != nil {
		return nil, err
	}

	return &page, nil
}
