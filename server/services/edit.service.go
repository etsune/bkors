package services

import (
	"context"
	"errors"
	"github.com/etsune/bkors/server/models"
	"github.com/etsune/bkors/server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"regexp"
	"strings"
	"time"
)

type EditService struct {
	col  *mongo.Collection
	ecol *mongo.Collection
	ctx  context.Context
}

func NewEditService(ctx context.Context, col, ecol *mongo.Collection) *EditService {
	return &EditService{col, ecol, ctx}
}

func (s *EditService) SetEditStatus(editIdStr string, status models.EditStatus) error {
	edit, err := s.Get(editIdStr)
	if err != nil {
		return err
	}

	edit.Status = status
	filter := bson.D{{"_id", edit.Id}}
	if _, err = s.col.ReplaceOne(s.ctx, filter, edit); err != nil {
		return err
	}

	return nil
}

func (s *EditService) Approve(editIdStr string) error {

	edit, err := s.Get(editIdStr)
	if err != nil {
		return err
	}

	var entry models.DBEntry
	if err = s.ecol.FindOne(s.ctx, bson.M{"_id": edit.EntryId}).Decode(&entry); err != nil {
		return err
	}

	entry.Entry.Hanja = edit.Result.Hanja
	entry.Entry.Hangul = edit.Result.Hangul
	entry.Entry.Transcription = edit.Result.Transcription
	entry.Entry.HomonymicNumber = edit.Result.HomonymicNumber
	entry.Entry.Body = edit.Result.Body
	entry.IsReviewed = edit.Result.IsReviewed
	entry.UpdatedAt = time.Now()

	entry.HeaderSearch = []string{}

	if len(entry.Entry.Hangul) > 0 {
		entry.HeaderSearch = append(entry.HeaderSearch, strings.Split(entry.Entry.Hangul, ",")...)
	}
	if len(entry.Entry.Hanja) > 0 {
		entry.HeaderSearch = append(entry.HeaderSearch, strings.Split(entry.Entry.Hanja, ",")...)
	}

	entry.BodySearch = strings.Join(entry.HeaderSearch, ",") + "\n" + entry.Entry.Body

	filter := bson.D{{"_id", entry.Id}}
	if _, err = s.ecol.ReplaceOne(s.ctx, filter, entry); err != nil {
		return err
	}

	edit.Status = models.StatusApproved
	filter = bson.D{{"_id", edit.Id}}
	if _, err = s.col.ReplaceOne(s.ctx, filter, edit); err != nil {
		return err
	}

	return nil
}

func (s *EditService) Get(editIdStr string) (*models.DBEdit, error) {
	editId, err := primitive.ObjectIDFromHex(editIdStr)
	if err != nil {
		return nil, err
	}
	var edit models.DBEdit
	if err = s.col.FindOne(s.ctx, bson.M{"_id": editId}).Decode(&edit); err != nil {
		return nil, err
	}

	return &edit, nil
}

func (s *EditService) GetAll(pending bool) (*[]models.DBEdit, error) {
	filter := bson.M{}
	if pending {
		filter = bson.M{"status": 0}
	}
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(100)
	cursor, err := s.col.Find(s.ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var res []models.DBEdit
	if err = cursor.All(s.ctx, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *EditService) CreateEdit(edit *models.EditEntry, entryId primitive.ObjectID, username string) (primitive.ObjectID, error) {

	var entry models.DBEntry
	err := s.ecol.FindOne(s.ctx, bson.M{"_id": entryId}).Decode(&entry)
	if err != nil {
		return primitive.NewObjectID(), err
	}

	oldedit := models.EditEntry{
		IsReviewed:      entry.IsReviewed,
		Hangul:          entry.Entry.Hangul,
		Hanja:           entry.Entry.Hanja,
		HomonymicNumber: entry.Entry.HomonymicNumber,
		Transcription:   entry.Entry.Transcription,
		Body:            entry.Entry.Body,
		Meta:            utils.GetMeta(&entry.Placement),
	}

	re := regexp.MustCompile(`\r*\n+`)

	edit.Hangul = strings.TrimSpace(replaceSymbols(edit.Hangul))
	edit.Hanja = strings.TrimSpace(replaceSymbols(edit.Hanja))
	edit.Transcription = strings.TrimSpace(replaceSymbols(edit.Transcription))
	edit.HomonymicNumber = strings.TrimSpace(edit.HomonymicNumber)
	edit.Body = strings.TrimSpace(replaceSymbols(edit.Body))
	edit.Body = re.ReplaceAllString(edit.Body, "\n")
	edit.Meta = oldedit.Meta

	if edit.Hangul == oldedit.Hangul &&
		edit.Hanja == oldedit.Hanja &&
		edit.Transcription == oldedit.Transcription &&
		edit.HomonymicNumber == oldedit.HomonymicNumber &&
		edit.Body == oldedit.Body {
		return primitive.NewObjectID(), errors.New("правка ничего не исправляет")
	}

	dbedit := models.DBEdit{
		Id:        primitive.NewObjectID(),
		EntryId:   entryId,
		Status:    models.StatusNew,
		Source:    oldedit,
		Result:    *edit,
		Author:    username,
		Approver:  "",
		Image:     entry.Image,
		CreatedAt: time.Now(),
	}

	res, err := s.col.InsertOne(s.ctx, dbedit)
	if err != nil {
		return primitive.NewObjectID(), err
	}

	return res.InsertedID.(primitive.ObjectID), nil
}

func replaceSymbols(txt string) string {
	txt = strings.ReplaceAll(txt, "о1", "ɔ")
	txt = strings.ReplaceAll(txt, "~=", "≅")
	txt = strings.ReplaceAll(txt, "<>", "◇")
	txt = strings.ReplaceAll(txt, "--", "—")

	txt = strings.ReplaceAll(txt, "〈", "<")
	txt = strings.ReplaceAll(txt, "＜", "<")
	txt = strings.ReplaceAll(txt, "～", "~")

	return txt
}
