package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-pg/pg/v10"
	"github.com/jasonlvhit/gocron"
	"github.com/joho/godotenv"
)

func ping(db *pg.DB) {
	var urls []PingURL
	err := db.Model(&urls).Select()

	if err != nil {
		panic(err)
	}
	fmt.Println(urls)

	for _, url := range urls {
		fmt.Println("Pinging:" + url.URL)
	}

	fmt.Println("Completed 1 cycle")
}

func executeCronJob(db *pg.DB) {
	period, err := strconv.ParseUint(os.Getenv("PERIOD"), 10, 64)
	if err != nil {
		// Set to default period if couldn't get value from dotenv
		period = 600
	}
	gocron.Every(period).Second().Do(ping, db)
	<-gocron.Start()
}

func main() {
	// Initialize environment
	err := godotenv.Load(getenv("ENV_FILE", "/vault/secrets/env"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := pg.Connect(&pg.Options{
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASS"),
		Addr:     os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"),
		Database: os.Getenv("POSTGRES_DB"),
	})

	defer db.Close()

	err = CreateSchema(db)

	if err != nil {
		log.Fatal(err)
	}

	log.Print("Connected to database!")

	go executeCronJob(db)

	log.Print("Started running cron job!")
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
