package services

import (
	"bufio"
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"strings"

	"github.com/etsune/bkors/server/models"
	"github.com/etsune/bkors/server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EntryService struct {
	col *mongo.Collection
	ctx context.Context
}

func NewEntryService(ctx context.Context, col *mongo.Collection) *EntryService {
	return &EntryService{col, ctx}
}

func (e *EntryService) GetEntriesForPage(volume, page int) (*[]models.DBEntry, error) {
	filter := bson.M{"placement.v": volume, "placement.pg": page}
	cursor, err := e.col.Find(e.ctx, filter, options.Find().SetLimit(1000))
	if err != nil {
		return nil, err
	}
	var res []models.DBEntry
	if err = cursor.All(e.ctx, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (e *EntryService) SearchEntries(term string) ([]models.DBEntry, error) {
	term = strings.TrimSpace(term)
	if len(term) == 0 {
		return []models.DBEntry{}, nil
	}
	filter1 := bson.M{"header_search": term}
	filter2 := bson.M{"body_search": bson.M{"$regex": term, "$options": "i"}}
	// filter2 := bson.M{"$text": bson.M{"$search": term}}
	cursor, err := e.col.Find(e.ctx, filter1, options.Find().SetLimit(100))
	if err != nil {
		return nil, err
	}

	var res []models.DBEntry
	var curEntry models.DBEntry
	existKeys := make(map[string]bool)

	for cursor.Next(e.ctx) {
		if err = cursor.Decode(&curEntry); err != nil {
			fmt.Println(err)
			break
		}
		if _, val := existKeys[curEntry.Id.Hex()]; !val {
			res = append(res, curEntry)
			existKeys[curEntry.Id.Hex()] = true
		}
	}

	newlimit := int64(100 - len(res))
	if newlimit < 10 {
		return res, nil
	}

	cursor, err = e.col.Find(e.ctx, filter2, options.Find().SetLimit(newlimit))
	if err != nil {
		return nil, err
	}

	for cursor.Next(e.ctx) {
		if err = cursor.Decode(&curEntry); err != nil {
			fmt.Println(err)
			break
		}
		if _, val := existKeys[curEntry.Id.Hex()]; !val {
			res = append(res, curEntry)
			existKeys[curEntry.Id.Hex()] = true
		}
	}

	return res, nil
}

func (e *EntryService) ExportEntries() string {
	col, _ := e.SearchEntries("")
	res := utils.GetPrevText()
	for _, entry := range col {
		res += utils.ConvertEntryToMultilineTxt(entry) + "\n\n"
	}
	return res
}

func (e *EntryService) ExportEntriesToTxt() error {
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"psort", 1}}).SetLimit(1000)

	resFile, err := os.Create("test.txt")
	if err != nil {
		return err
	}
	defer resFile.Close()

	resFile.WriteString(utils.GetPrevText() + "\n")

	var counter int64 = 0
	for {
		opts = opts.SetSkip(counter * 1000)
		cursor, err := e.col.Find(e.ctx, filter, opts)
		if err != nil {
			return err
		}
		var results []models.DBEntry
		if err = cursor.All(e.ctx, &results); err != nil {
			return err
		}
		if len(results) == 0 {
			break
		}
		currentText := ""
		for _, entry := range results {
			currentText += utils.ConvertEntryToMultilineTxt(entry) + "\n\n"
		}
		resFile.WriteString(currentText)
		counter++
	}

	return nil
}

func (e *EntryService) ImportFile(file multipart.File) error {
	scanner := bufio.NewScanner(file)
	var curEntry = models.DBEntry{}
	// var res []models.DBEntry
	ln := 0
	createNewEntry := false
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if len(text) == 0 {
			if ln >= 2 {
				curEntry.Entry.Body = strings.TrimSpace(curEntry.Entry.Body)
				curEntry.BodySearch = strings.Join(curEntry.HeaderSearch, ",") + "\n" + curEntry.Entry.Body
				if createNewEntry {
					curEntry.Id = primitive.NewObjectID()
					e.col.InsertOne(e.ctx, curEntry)
				} else {
					filter := bson.D{{"_id", curEntry.Id}}
					e.col.ReplaceOne(e.ctx, filter, curEntry)
				}
				createNewEntry = false
			}
			ln = 0
			curEntry = models.DBEntry{}
			continue
		}
		if text[0] == '#' {
			ln = 0
			continue
		}
		if ln == 0 {
			split, err := utils.ParseDataline(text, 8)
			if err != nil {
				continue
			}

			curEntry.IsReviewed = false
			if split[0] == "+" {
				curEntry.IsReviewed = true
			}
			curEntry.Placement.Volume = utils.ConvStrToInt(split[1])
			curEntry.Placement.Page = utils.ConvStrToInt(split[2])
			curEntry.Placement.Side = utils.ConvStrToInt(split[3])
			curEntry.Placement.Paragraph = utils.ConvStrToInt(split[4])
			curEntry.Placement.Coords = strings.TrimSpace(split[5])

			if split[6] == "aaa" {
				curEntry.Id = primitive.NewObjectID()
				createNewEntry = true
			} else {
				curEntry.Id, err = primitive.ObjectIDFromHex(split[6])
				if err != nil {
					continue
				}
			}
			curEntry.Image = strings.TrimSpace(split[7])
			curEntry.PlacementSort = utils.GetSortNum(curEntry.Placement)

		}
		if ln == 1 {
			split, err := utils.ParseDataline(text, 4)
			if err != nil {
				continue
			}
			curEntry.Entry.Hangul = split[0]
			curEntry.Entry.Hanja = split[1]
			curEntry.Entry.HomonymicNumber = split[2]
			curEntry.Entry.Transcription = split[3]
			curEntry.HeaderSearch = []string{}

			if len(curEntry.Entry.Hangul) > 0 {
				curEntry.HeaderSearch = append(curEntry.HeaderSearch, strings.Split(curEntry.Entry.Hangul, ",")...)
			}
			if len(curEntry.Entry.Hanja) > 0 {
				curEntry.HeaderSearch = append(curEntry.HeaderSearch, strings.Split(curEntry.Entry.Hanja, ",")...)
			}
		}
		if ln >= 2 {
			curEntry.Entry.Body += text + "\n"
		}
		ln++
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
