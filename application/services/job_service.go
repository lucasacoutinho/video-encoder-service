package services

import (
	"errors"
	"os"
	"strconv"

	"github.com/lucasacoutinho/video-encoder-service/application/repositories"
	"github.com/lucasacoutinho/video-encoder-service/domain"
)

type JobService struct {
	Job           *domain.Job
	JobRepository repositories.JobRepository
	VideoService  VideoService
}

func (j *JobService) Start() error {
	err := j.changeJobStatus("DOWNLOADING")
	if err != nil {
		return j.failJob(err)
	}
	err = j.VideoService.Download(os.Getenv("STORAGE_BUCKET"))
	if err != nil {
		return j.failJob(err)
	}

	err = j.changeJobStatus("FRAGMENTING")
	if err != nil {
		return j.failJob(err)
	}
	err = j.VideoService.Fragment()
	if err != nil {
		return j.failJob(err)
	}

	err = j.changeJobStatus("ENCODING")
	if err != nil {
		return j.failJob(err)
	}
	err = j.VideoService.Encode()
	if err != nil {
		return j.failJob(err)
	}

	err = j.performUpload()
	if err != nil {
		return j.failJob(err)
	}

	err = j.changeJobStatus("FINISHING")
	if err != nil {
		return j.failJob(err)
	}
	err = j.VideoService.Finsh()
	if err != nil {
		return j.failJob(err)
	}

	err = j.changeJobStatus("COMPLETED")
	if err != nil {
		return j.failJob(err)
	}

	return nil
}

func (j *JobService) performUpload() error {
	err := j.changeJobStatus("UPLOADING")
	if err != nil {
		return j.failJob(err)
	}

	videoUpload := NewVideoUpload()
	videoUpload.OutputBucket = os.Getenv("STORAGE_BUCKET")
	videoUpload.VideoPath = os.Getenv("STORAGE_LOCAL_PATH") + "/" + j.VideoService.Video.ID
	concurrency, _ := strconv.Atoi(os.Getenv("STORAGE_UPLOAD_CONCURRENCY"))
	done := make(chan string)

	go videoUpload.ProcessUpload(concurrency, done)

	result := <-done
	if result != "upload completed" {
		return j.failJob(errors.New(result))
	}

	return err
}

func (j *JobService) changeJobStatus(status string) error {
	var err error

	j.Job.Status = status
	j.Job, err = j.JobRepository.Update(j.Job)

	if err != nil {
		return j.failJob(err)
	}

	return nil
}

func (j *JobService) failJob(e error) error {
	j.Job.Status = "FAILED"
	j.Job.Error = e.Error()

	_, err := j.JobRepository.Update(j.Job)

	if err != nil {
		return err
	}

	return e
}
