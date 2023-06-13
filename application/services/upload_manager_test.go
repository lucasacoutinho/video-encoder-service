package services_test

import (
	"os"
	"testing"

	"github.com/lucasacoutinho/video-encoder-service/application/services"
	"github.com/stretchr/testify/require"
)

func TestVideoServiceUpload(t *testing.T) {
	video, videoRepo := prepare()

	videoService := services.NewVideoService()
	videoService.Video = video
	videoService.VideoRepository = videoRepo

	bucket := os.Getenv("STORAGE_BUCKET")

	err := videoService.Download(bucket)
	require.Nil(t, err)

	err = videoService.Fragment()
	require.Nil(t, err)

	err = videoService.Encode()
	require.Nil(t, err)

	videoUpload := services.NewVideoUpload()
	videoUpload.OutputBucket = bucket
	videoUpload.VideoPath = os.Getenv("STORAGE_LOCAL_PATH") + "/" + video.ID

	done := make(chan string)
	concurrency := 50
	go videoUpload.ProcessUpload(concurrency, done)

	result := <-done
	require.Equal(t, result, "upload completed")

	err = videoService.Finsh()
	require.Nil(t, err)
}
