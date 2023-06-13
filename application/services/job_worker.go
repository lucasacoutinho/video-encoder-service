package services

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lucasacoutinho/video-encoder-service/domain"
	"github.com/lucasacoutinho/video-encoder-service/framework/utils"
	"github.com/streadway/amqp"
)

type JobWorkerResult struct {
	Job     domain.Job
	Message *amqp.Delivery
	Error   error
}

var Mutex = &sync.Mutex{}

func JobWorker(
	workerID int,
	job domain.Job,
	jobService JobService,
	messagesChannel chan amqp.Delivery,
	resultChannel chan JobWorkerResult,
) {
	for message := range messagesChannel {
		err := utils.IsJson(string(message.Body))
		if err != nil {
			resultChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		Mutex.Lock()
		err = json.Unmarshal(message.Body, &jobService.VideoService.Video)
		jobService.VideoService.Video.ID = uuid.NewString()
		Mutex.Unlock()
		if err != nil {
			resultChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		err = jobService.VideoService.Video.Validate()
		if err != nil {
			resultChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		Mutex.Lock()
		err = jobService.VideoService.InsertVideo()
		Mutex.Unlock()
		if err != nil {
			resultChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		job.ID = uuid.NewString()
		job.Video = jobService.VideoService.Video
		job.OutputBucketPath = os.Getenv("STORAGE_BUCKET")
		job.Status = "STARTING"
		job.CreatedAt = time.Now()

		Mutex.Lock()
		_, err = jobService.JobRepository.Insert(&job)
		Mutex.Unlock()
		if err != nil {
			resultChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		jobService.Job = &job

		err = jobService.Start()
		if err != nil {
			resultChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		resultChannel <- returnJobResult(job, message, nil)
	}
}

func returnJobResult(job domain.Job, message amqp.Delivery, err error) JobWorkerResult {
	result := JobWorkerResult{
		Job:     job,
		Message: &message,
		Error:   err,
	}
	return result
}
