package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/gorilla/mux"
	"github.com/jasonlvhit/gocron"
	"github.com/joho/godotenv"
)

// Ping - Model of the URL to be pinged
type Ping struct {
	ID          int64
	Name        string    `sql:",unique,notnull"`
	URL         string    `sql:",notnull"`
	Status      string    `sql:",notnull"`
	Count       int64     `sql:",notnull"`
	PassedCount int64     `sql:",notnull"`
	LastChecked time.Time `sql:",notnull"`
	CreatedDate time.Time `sql:",notnull"`
}

func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*Ping)(nil),
		// For more models here.
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			// Temp: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// App tmux application
type App struct {
	Router *mux.Router
	DB     *pg.DB
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func ping(db *pg.DB) {
	var urls []Ping
	err := db.Model(&urls).Select()

	if err != nil {
		panic(err)
	}

	fmt.Println(urls)

	for _, url := range urls {
		fmt.Printf("Pinging: %s, count: %d", url.URL, (url.Count + 1))
		res, err := http.Get(url.URL)

		if err != nil {
			log.Fatalln(err)
		}

		pingURL := url
		pingURL.Count = pingURL.Count + 1

		if res.StatusCode == 200 {
			pingURL.Status = "green"
			pingURL.PassedCount = pingURL.PassedCount + 1
		}

		if res.StatusCode != 200 && url.Status == "green" {
			pingURL.Status = "red"
		}

		pingURL.LastChecked = time.Now()

		db.Model(&pingURL).WherePK().Update()
	}
}

func executeCronJob(db *pg.DB) {
	period, err := strconv.ParseUint(os.Getenv("PERIOD"), 10, 64)
	if err != nil {
		// Set to default period if couldn't get value from dotenv
		period = 5
	}
	gocron.Every(period).Second().Do(ping, db)
	<-gocron.Start()
}

// Initialize initializing the application
func (app *App) Initialize(dbOptions *pg.Options) {
	app.DB = pg.Connect(dbOptions)
	app.Router = mux.NewRouter()
	app.initializeRoutes()
	createSchema(app.DB)
}

// Run start running app
func (app *App) Run(addr string) {
	go executeCronJob(app.DB)
	log.Print("Started running cron job!")
	log.Fatal(http.ListenAndServe(":8080", app.Router))
}

func main() {
	// Initialize environment
	err := godotenv.Load(getenv("ENV_FILE", "/vault/secrets/env"))

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbOptions := &pg.Options{
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASS"),
		Addr:     os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"),
		Database: os.Getenv("POSTGRES_DB"),
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Print("Connected to database!")

	app := App{}
	app.Initialize(dbOptions)
	log.Print("Initialized app!")

	app.Run(":8080")

	defer app.DB.Close()
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func getPingUrls(db *pg.DB) ([]Ping, error) {

	var urls []Ping

	err := db.Model(&urls).Select()

	if err != nil {
		panic(err)
	}

	fmt.Println(urls)

	return urls, nil
}

func (app *App) getPingUrls(w http.ResponseWriter, r *http.Request) {

	urls, err := getPingUrls(app.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, urls)
}

func (app *App) getPingURLs(w http.ResponseWriter, r *http.Request) {

	urls, err := getPingUrls(app.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, urls)
}

func (url *Ping) createPingURL(db *pg.DB) error {
	url.CreatedDate = time.Now()
	_, err := db.Model(url).Insert()

	if err != nil {
		return err
	}
	return nil
}

func (app *App) createPingURL(w http.ResponseWriter, r *http.Request) {
	var url Ping
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&url); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()
	err := url.createPingURL(app.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, url)
}

func (app *App) initializeRoutes() {
	app.Router.HandleFunc("/ping", app.getPingUrls).Methods("GET")
	app.Router.HandleFunc("/ping", app.createPingURL).Methods("POST")
}
