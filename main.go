package main

import (
	"diary_api/controller"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	loadDatabase()
	serveApplication()
}

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func loadDatabase() {
	// database.Connect()
	// database.Database.AutoMigrate(&model.User{})
	// database.Database.AutoMigrate(&model.Entry{})
}

func serveApplication() {
	router := gin.Default()

	apiRoutes := router.Group("/api")
	apiRoutes.POST("/mails/create", controller.CreateMail)
	apiRoutes.GET("/mails/list", controller.GetMails)

	// apiRoutes.POST("/storage/upload", controller.Upload)

	router.Run(":8888")
	fmt.Println("Server running on port 8888")
}