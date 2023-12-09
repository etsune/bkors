package services

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"github.com/etsune/bkors/server/models"
	"github.com/etsune/bkors/server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

type DownloadService struct {
	dcol *mongo.Collection
	ecol *mongo.Collection
	ctx  context.Context
}

func NewDownloadService(ctx context.Context, dcol *mongo.Collection, ecol *mongo.Collection) *DownloadService {
	return &DownloadService{
		dcol, ecol, ctx,
	}
}

func (s *DownloadService) ExportAll(filename, path string, currentTime time.Time) error {
	// remove old files
	if _, err := s.dcol.DeleteMany(s.ctx, bson.M{}); err != nil {
		//return err
	}
	if err := os.RemoveAll(path); err != nil {
		//return err
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		//return err
	}

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"psort", 1}}).SetLimit(1000)

	var buff bytes.Buffer
	zipW := zip.NewWriter(&buff)
	f, err := zipW.Create(filename)
	if err != nil {
		return nil
	}

	//resFile, err := os.Create(path + filename)
	//if err != nil {
	//	return err
	//}
	//defer resFile.Close()

	//if _, err = resFile.WriteString(utils.GetPrevText() + "\n"); err != nil {
	//	return err
	//}

	_, err = f.Write([]byte(utils.GetPrevText() + "\n"))
	if err != nil {
		return nil
	}

	var counter int64 = 0
	for {
		opts = opts.SetSkip(counter * 1000)
		cursor, err := s.ecol.Find(s.ctx, filter, opts)
		if err != nil {
			return err
		}
		var results []models.DBEntry
		if err = cursor.All(s.ctx, &results); err != nil {
			return err
		}
		if len(results) == 0 {
			break
		}
		currentText := ""
		for _, entry := range results {
			currentText += utils.ConvertEntryToMultilineTxt(entry) + "\n\n"
		}
		//if _, err = resFile.WriteString(currentText); err != nil {
		//	return err
		//}
		_, err = f.Write([]byte(currentText))
		if err != nil {
			return nil
		}
		counter++
	}

	err = zipW.Close()
	if err != nil {
		return nil
	}

	err = os.WriteFile(path+filename+".zip", buff.Bytes(), os.ModePerm)
	if err != nil {
		return nil
	}

	var fileSize string
	fi, err := os.Stat(path + filename + ".zip")
	if err != nil {
		fileSize = "-"
	} else {
		fileSize = fmt.Sprintf("%.2f Мб", float64(fi.Size())/1024/1024)
	}

	dbdl := models.DBDownload{
		Id:       primitive.NewObjectID(),
		Filename: filename + ".zip",
		Path:     path,
		Time:     currentTime.Format("2006-01-02 15:04:05"),
		Size:     fileSize,
	}

	if _, err = s.dcol.InsertOne(s.ctx, dbdl); err != nil {
		return err
	}

	return nil
}

func (s *DownloadService) GetAll() (*[]models.DBDownload, error) {
	cursor, err := s.dcol.Find(s.ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var res []models.DBDownload
	if err = cursor.All(s.ctx, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
