package main

import (
	"context"
	"fmt"
	"time"

	"github.com/etsune/bkors/server/config"
	"github.com/etsune/bkors/server/services"

	// echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		fmt.Println("Could not load environment variables", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())

	dbClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.DBUri))
	if err != nil {
		panic(err)
	}

	if err := dbClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	fmt.Println("MongoDB successfully connected...")

	// e.Use(echojwt.JWT([]byte(config.AccessTokenPrivateKey)))

	handler := &AppHandler{
		entryService:    services.NewEntryService(context.TODO(), dbClient.Database("bkors").Collection("entries")),
		authService:     services.NewAuthService(config),
		userService:     services.NewUserService(context.TODO(), dbClient.Database("bkors").Collection("users"), config),
		sheetService:    services.NewSheetService(context.TODO(), dbClient.Database("bkors").Collection("pages")),
		editService:     services.NewEditService(context.TODO(), dbClient.Database("bkors").Collection("edits"), dbClient.Database("bkors").Collection("entries")),
		downloadService: services.NewDownloadService(context.TODO(), dbClient.Database("bkors").Collection("downloads"), dbClient.Database("bkors").Collection("entries")),
	}

	e.Use(SetRequestUser(handler.userService, &config))

	router(e, handler)
	e.Static("/static", "assets")

	go exportJob(config.ExportDir, handler.downloadService)

	e.Logger.Fatal(e.Start(":" + config.Port))
}

func exportJob(exportDir string, e *services.DownloadService) {
	fmt.Println("Starting export job")
	startExport(exportDir, e)
	for range time.Tick(12 * time.Hour) {
		startExport(exportDir, e)
	}
}

func startExport(exportDir string, e *services.DownloadService) {
	currentTime := time.Now()
	fmt.Println("Exporting ")
	filename := fmt.Sprintf("bkors-export-%s.txt", currentTime.Format("2006-01-02_15-04-05"))
	_ = e.ExportAll(filename, exportDir, currentTime)
}
