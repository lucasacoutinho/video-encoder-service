package main

import (
	"log"
	"os"
	"strconv"

	"github.com/lucasacoutinho/video-encoder-service/application/services"
	"github.com/lucasacoutinho/video-encoder-service/framework/database"
	"github.com/lucasacoutinho/video-encoder-service/framework/queue"
	"github.com/streadway/amqp"
)

var DB database.Database

func init() {
	autoMigrateDB, err := strconv.ParseBool(os.Getenv("AUTO_MIGRATE_DB"))
	if err != nil {
		log.Fatalf("Error parsing env var: AUTO_MIGRATE_DB")
	}

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		log.Fatalf("Error parsing env var: DEBUG")
	}

	DB.AutoMigrateDB = autoMigrateDB
	DB.Debug = debug
	DB.DSN = os.Getenv("DSN")
	DB.DBType = os.Getenv("DB_TYPE")
	DB.Env = os.Getenv("ENV")
}

func main() {
	conn, err := DB.Connect()
	if err != nil {
		log.Fatalf("Error connecting to DB")
	}
	defer conn.Close()

	rabbitMQ := queue.NewRabbitMQ()
	ch := rabbitMQ.Connect()
	defer ch.Close()

	messageChannel := make(chan amqp.Delivery)
	rabbitMQ.Consume(messageChannel)

	jobReturnChannel := make(chan services.JobWorkerResult)

	jobManager := services.NewJobManager(conn, rabbitMQ, messageChannel, jobReturnChannel)
	jobManager.Start(ch)
}
