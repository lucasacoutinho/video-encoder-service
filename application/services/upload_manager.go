package services

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type VideoUpload struct {
	Paths        []string
	VideoPath    string
	OutputBucket string
	Errors       []string
}

func NewVideoUpload() *VideoUpload {
	return &VideoUpload{}
}

func (v *VideoUpload) UploadObject(path string, client *minio.Client, ctx context.Context) error {
	f := strings.Split(path, os.Getenv("STORAGE_LOCAL_PATH")+"/")

	if _, err := client.FPutObject(
		ctx,
		v.OutputBucket,
		f[1],
		path,
		minio.PutObjectOptions{},
	); err != nil {
		log.Printf("Error uploading file: %s", f[1])
		log.Println(err)
		return err
	}

	return nil
}

func (v *VideoUpload) loadPaths() error {
	err := filepath.Walk(v.VideoPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			v.Paths = append(v.Paths, path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (v *VideoUpload) ProcessUpload(concurrency int, done chan string) error {
	err := v.loadPaths()

	if err != nil {
		return err
	}

	client, ctx, err := getClientUpload()
	if err != nil {
		return err
	}

	in := make(chan int, runtime.NumCPU())
	ret := make(chan string)
	for process := 0; process < concurrency; process++ {
		go v.uploadWorker(in, ret, client, ctx)
	}

	go func() {
		for x := 0; x < len(v.Paths); x++ {
			in <- x
		}
		close(in)
	}()

	for r := range ret {
		if r != "" {
			done <- r
			break
		}
	}

	return nil
}

func (v *VideoUpload) uploadWorker(in chan int, ret chan string, client *minio.Client, ctx context.Context) {
	for x := range in {
		err := v.UploadObject(v.Paths[x], client, ctx)

		if err != nil {
			v.Errors = append(v.Errors, v.Paths[x])
			log.Printf("error during the upload: %v. Error: %v", v.Paths[x], err)
			ret <- err.Error()
		}

		ret <- ""
	}

	ret <- "upload completed"
}

func getClientUpload() (*minio.Client, context.Context, error) {
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
		return nil, nil, err
	}

	return client, ctx, nil
}
