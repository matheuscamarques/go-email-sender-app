package main

import (
	"go-email-sender-app/jsonsender"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	SMTPServer = "smtp.gmail.com"
	SMTPPort   = 587
)

func logInit() func() error {
	LOG_FILE := "./log"

	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
		return nil
	}

	log.SetOutput(logFile)

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	return logFile.Close
}

func getCredentials() (username string, password string) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	username = os.Getenv("EMAIL")

	password = os.Getenv("PASSWORD")

	return
}

func main() {
	close := logInit()
	defer close()

	username, password := getCredentials()

	log.Println("Starting...")
	js, err := jsonsender.New(SMTPServer, SMTPPort, username, password)
	if err != nil {
		log.Println("[ERROR]", err)
		panic(err)
	}

	log.Println("[INFO] Loading JSON...")
	messages, err := js.GetJsonFile("./data.json")
	if err != nil {
		log.Println("[ERROR]", err)
		panic(err)
	}

	log.Println("[INFO] Sending messages...")
	err = js.Send(messages...)
	if err != nil {
		log.Println("[ERROR]", err)
		panic(err)
	}

	log.Println("[INFO] Done!")
}
