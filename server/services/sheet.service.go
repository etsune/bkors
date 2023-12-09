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

func (s *SheetService) Get(dict string, vol, pg int) (*models.DBPage, error) {
	if vol < 0 || pg < 0 {
		return nil, errors.New("incorrect value")
	}
	if vol > 10000 || pg > 10000 {
		return nil, errors.New("incorrect value")
	}

	filter := bson.M{"dict": dict, "vol": vol, "p": pg}
	var page models.DBPage
	if err := s.col.FindOne(s.ctx, filter).Decode(&page); err != nil {
		return nil, err
	}

	return &page, nil
}

func (s *SheetService) GetByNum(dict string, num int) (*models.DBPage, error) {
	if num < 0 {
		return nil, errors.New("incorrect value")
	}
	if num > 10000 {
		return nil, errors.New("incorrect value")
	}

	filter := bson.M{"dict": dict, "num": num}
	var page models.DBPage
	if err := s.col.FindOne(s.ctx, filter).Decode(&page); err != nil {
		return nil, err
	}

	return &page, nil
}
