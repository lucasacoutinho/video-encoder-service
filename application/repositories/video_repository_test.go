package repositories_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lucasacoutinho/video-encoder-service/application/repositories"
	"github.com/lucasacoutinho/video-encoder-service/domain"
	"github.com/lucasacoutinho/video-encoder-service/framework/database"
	"github.com/stretchr/testify/require"
)

func TestVideoRepositoryDBInsert(t *testing.T) {
	db := database.NewDBTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.New().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repo := repositories.NewVideoRepository(db)
	repo.Insert(video)

	v, err := repo.Find(video.ID)

	require.NotEmpty(t, v.ID)
	require.Nil(t, err)
	require.Equal(t, v.ID, video.ID)
}
