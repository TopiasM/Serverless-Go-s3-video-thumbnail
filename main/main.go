package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"io"
	"io/ioutil"
	"strings"
	"bytes"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func Handler(ctx context.Context, s3Event events.S3Event) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		log.Print(err)
	}

	svc := s3.New(sess)

	vidKey := s3Event.Records[0].S3.Object.Key

	result, err := svc.GetObject(&s3.GetObjectInput {
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key: aws.String(vidKey),
	})
	if err != nil {
		log.Print(err)
	}

	log.Printf("%s to thumbnail", vidKey)

	tmpVidPath := strings.Join([]string("/tmp/video", os.GetEnv("VIDEO_EXTENSION")))
	out, err := os.Create(tmpVidPath)
	defer out.Close()

	resp := result
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)

	path, _ := filepath.Abs("./bin/ffmpeg");
	tmpThumbPath := strings.Join([]string{"/tmp/thumb", os.Getenv("THUMB_EXTENSION")}, "")
	cmd := exec.Command(path, "-i", tmpVidPath, "-movflags", "faststart" ,"-ss", os.Getenv("SCREENSHOT_TIME"), "-vframes", "1", "-s", os.Getenv("RESOLUTION"), tmpThumbPath)
	err = cmd.Run()
	if err != nil {
		log.Print(err)
	}
	defer os.Remove(tmpThumbPath)

	nameSlice := strings.Split(vidKey, "/")
	typeSlice := strings.Split(nameSlice[1], ".")
	thumbKey := strings.Join([]string{nameSlice[0], "-thumb/", typeSlice[0], os.Getenv("THUMB_EXTENSION")}, "")
	dat, _ := ioutil.ReadFile(tmpThumbPath)

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key: aws.String(thumbKey),
		Body: bytes.NewReader(dat),
	})
	if(err != nil) {
		log.Printf("%s", err)
	}
}

func main() {
	lambda.Start(Handler)
}
