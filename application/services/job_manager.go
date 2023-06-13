package services

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/lucasacoutinho/video-encoder-service/application/repositories"
	"github.com/lucasacoutinho/video-encoder-service/domain"
	"github.com/lucasacoutinho/video-encoder-service/framework/queue"
	"github.com/streadway/amqp"
)

type JobManager struct {
	DB               *gorm.DB
	Job              domain.Job
	MessageChannel   chan amqp.Delivery
	JobReturnChannel chan JobWorkerResult
	RabbitMQ         *queue.RabbitMQ
}

type JobNotificationError struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func NewJobManager(
	db *gorm.DB,
	rabbitMQ *queue.RabbitMQ,
	messageChannel chan amqp.Delivery,
	jobReturnChannel chan JobWorkerResult,
) *JobManager {
	return &JobManager{
		DB:               db,
		RabbitMQ:         rabbitMQ,
		MessageChannel:   messageChannel,
		JobReturnChannel: jobReturnChannel,
		Job:              domain.Job{},
	}
}

func (j *JobManager) Start(ch *amqp.Channel) {
	videoService := NewVideoService()
	videoService.VideoRepository = repositories.NewVideoRepository(j.DB)

	jobService := JobService{
		JobRepository: repositories.NewJobRepository(j.DB),
		VideoService:  videoService,
	}

	concurrency, err := strconv.Atoi(os.Getenv("CONCURRENCY_WORKERS"))
	if err != nil {
		log.Fatalf("Error loading var from enviroment: CONCURRENCY_WORKERS")
	}

	for process := 1; process <= concurrency; process++ {
		go JobWorker(
			process,
			j.Job,
			jobService,
			j.MessageChannel,
			j.JobReturnChannel,
		)
	}

	for result := range j.JobReturnChannel {
		if result.Error != nil {
			err = j.checkParseErrors(result)
		} else {
			err = j.notifySuccess(result, ch)
		}

		if err != nil {
			result.Message.Reject(false)
		}
	}
}

func (j *JobManager) notifySuccess(result JobWorkerResult, ch *amqp.Channel) error {
	Mutex.Lock()
	json, err := json.Marshal(result.Job)
	Mutex.Unlock()
	if err != nil {
		return err
	}

	err = j.notify(json)
	if err != nil {
		return err
	}

	err = result.Message.Ack(false)
	if err != nil {
		return err
	}

	return nil
}

func (j *JobManager) checkParseErrors(result JobWorkerResult) error {
	if result.Job.ID != "" {
		log.Printf("MessageID: %v. Error during job: %v with video: %v. Error: %v",
			result.Message.DeliveryTag,
			result.Job.ID,
			result.Job.Video.ID,
			result.Error.Error(),
		)
	} else {
		log.Printf("MessageID: %v. Error parsing message: %v",
			result.Message.DeliveryTag,
			result.Error,
		)
	}

	errMsg := JobNotificationError{
		Message: string(result.Message.Body),
		Error:   result.Error.Error(),
	}

	json, err := json.Marshal(errMsg)
	if err != nil {
		return err
	}

	err = j.notify(json)
	if err != nil {
		return err
	}

	err = result.Message.Reject(false)
	if err != nil {
		return err
	}

	return nil
}

func (j *JobManager) notify(json []byte) error {
	err := j.RabbitMQ.Notify(
		string(json),
		"application/json",
		os.Getenv("RABBITMQ_NOTIFICATION_EX"),
		os.Getenv("RABBITMQ_NOTIFICATION_ROUTING_KEY"),
	)

	if err != nil {
		return err
	}

	return nil
}
