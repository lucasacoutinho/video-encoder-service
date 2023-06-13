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

func TestJobRepositoryDBInsert(t *testing.T) {
	db := database.NewDBTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.New().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repoVideo := repositories.NewVideoRepository(db)
	repoVideo.Insert(video)

	job, err := domain.NewJob("output_path", "pending", video)
	require.Nil(t, err)

	repoJob := repositories.NewJobRepository(db)
	repoJob.Insert(job)

	j, err := repoJob.Find(job.ID)
	require.Nil(t, err)
	require.NotEmpty(t, j.ID)
	require.Equal(t, j.ID, job.ID)
	require.Equal(t, j.Video.ID, video.ID)
}

func TestJobRepositoryDBUpdate(t *testing.T) {
	db := database.NewDBTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.New().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repoVideo := repositories.NewVideoRepository(db)
	repoVideo.Insert(video)

	job, err := domain.NewJob("output_path", "pending", video)
	require.Nil(t, err)

	repoJob := repositories.NewJobRepository(db)
	repoJob.Insert(job)

	job.Status = "Completed"

	repoJob.Update(job)

	j, err := repoJob.Find(job.ID)
	require.Nil(t, err)
	require.NotEmpty(t, j.ID)
	require.Equal(t, j.Status, job.Status)
}
