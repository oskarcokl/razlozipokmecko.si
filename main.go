package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/oskarcokl/razlozipokmecko.si/db"
	"github.com/oskarcokl/razlozipokmecko.si/handlers"
	"github.com/oskarcokl/razlozipokmecko.si/services"
)


type Page struct {
    Title string
    Body []byte
}


func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    uri := os.Getenv("MONGODB_URI")

    ms := db.NewMongoConnection(uri)
    ps := services.NewPageService(ms)
    h := handlers.New(ps)
    h.ServeHTTP()
}