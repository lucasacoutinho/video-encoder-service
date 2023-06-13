package services_test

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lucasacoutinho/video-encoder-service/application/repositories"
	"github.com/lucasacoutinho/video-encoder-service/application/services"
	"github.com/lucasacoutinho/video-encoder-service/domain"
	"github.com/lucasacoutinho/video-encoder-service/framework/database"
	"github.com/stretchr/testify/require"
)

func prepare() (*domain.Video, *repositories.VideoRepositoryDB) {
	db := database.NewDBTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.New().String()
	video.FilePath = "1.mp4"
	video.CreatedAt = time.Now()

	repo := repositories.NewVideoRepository(db)

	return video, repo
}

func TestVideoServiceDownload(t *testing.T) {
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

	err = videoService.Finsh()
	require.Nil(t, err)
}
