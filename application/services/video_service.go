package services

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/lucasacoutinho/video-encoder-service/application/repositories"
	"github.com/lucasacoutinho/video-encoder-service/domain"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.VideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func (v *VideoService) InsertVideo() error {
	_, err := v.VideoRepository.Insert(v.Video)
	if err != nil {
		return err
	}

	return nil
}

func (v *VideoService) Download(bucket string) error {
	ctx := context.Background()

	client, err := minio.New(
		os.Getenv("STORAGE_ENPOINT"),
		&minio.Options{
			Creds: credentials.NewStaticV4(
				os.Getenv("STORAGE_ACCESS_KEY"),
				os.Getenv("STORAGE_SECRET_KEY"),
				"",
			),
			Secure: false,
		})

	if err != nil {
		return err
	}

	object, err := client.GetObject(ctx, bucket, v.Video.FilePath, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer object.Close()

	body, err := ioutil.ReadAll(object)
	if err != nil {
		return err
	}

	file, err := os.Create(os.Getenv("STORAGE_LOCAL_PATH") + "/" + v.Video.ID + ".mp4")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(body)
	if err != nil {
		return err
	}

	log.Printf("video %v has been stored", v.Video.ID)

	return nil
}

func (v *VideoService) Fragment() error {
	err := os.Mkdir(os.Getenv("STORAGE_LOCAL_PATH")+"/"+v.Video.ID, os.ModePerm)

	if err != nil {
		return err
	}

	source := os.Getenv("STORAGE_LOCAL_PATH") + "/" + v.Video.ID + ".mp4"
	target := os.Getenv("STORAGE_LOCAL_PATH") + "/" + v.Video.ID + ".frag"

	cmd := exec.Command("mp4fragment", source, target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(out)

	return nil
}

func (v *VideoService) Encode() error {
	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, os.Getenv("STORAGE_LOCAL_PATH")+"/"+v.Video.ID+".frag")
	cmdArgs = append(cmdArgs, "--use-segment-timeline")
	cmdArgs = append(cmdArgs, "--o")
	cmdArgs = append(cmdArgs, os.Getenv("STORAGE_LOCAL_PATH")+"/"+v.Video.ID)
	cmdArgs = append(cmdArgs, "-f")
	cmdArgs = append(cmdArgs, "--exec-dir")
	cmdArgs = append(cmdArgs, "/opt/bento4/bin/")

	cmd := exec.Command("mp4dash", cmdArgs...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(out)

	return nil
}

func (v *VideoService) Finsh() error {
	err := os.Remove(os.Getenv("STORAGE_LOCAL_PATH") + "/" + v.Video.ID + ".mp4")
	if err != nil {
		log.Println("error removing mp4 ", v.Video.ID, ".mp4")
		return err
	}

	err = os.Remove(os.Getenv("STORAGE_LOCAL_PATH") + "/" + v.Video.ID + ".frag")
	if err != nil {
		log.Println("error removing frag", v.Video.ID, ".frag")
		return err
	}

	err = os.RemoveAll(os.Getenv("STORAGE_LOCAL_PATH") + "/" + v.Video.ID)
	if err != nil {
		log.Println("error removing dir", v.Video.ID)
		return err
	}

	return nil
}

func printOutput(out []byte) {
	if len(out) > 0 {
		log.Printf("============> Output: %s\n", string(out))
	}
}
