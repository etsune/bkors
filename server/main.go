package main

import (
	"context"
	"fmt"

	"github.com/etsune/bkors/config"
	"github.com/etsune/bkors/services"

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
		entryService: services.NewEntryService(context.TODO(), dbClient.Database("bkors").Collection("entries")),
		authService:  services.NewAuthService(config),
		userService:  services.NewUserService(context.TODO(), dbClient.Database("bkors").Collection("users"), config),
	}

	e.Use(SetRequestUser(handler.userService, &config))

	router(e, handler)
	e.Static("/static", "assets")

	// sch := gocron.NewScheduler(time.UTC)
	// sch.Every(1).Hours().Do(handler.entryService.ExportEntriesToTxt)
	// sch.StartAsync()

	// handler.entryService.ExportEntriesToTxt()

	e.Logger.Fatal(e.Start("127.0.0.1:" + config.Port))
}
