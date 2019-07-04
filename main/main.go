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
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})

	svc := s3.New(sess)

	vidKey := s3Event.Records[0].S3.Object.Key

	result, err := svc.GetObject(&s3.GetObjectInput {
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key: aws.String(vidKey),
	})

	log.Printf("%s to thumb", vidKey)

	out, err := os.Create("/tmp/vid.mp4")
	defer out.Close()

	resp := result
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)

	path, _ := filepath.Abs("./bin/ffmpeg");
	cmd := exec.Command(path, "-i", "/tmp/vid.mp4", "-movflags", "faststart" ,"-ss", os.Getenv("SCREENSHOT_TIME"), "-vframes", "1", "-s", os.Getenv("RESOLUTION"), "/tmp/thumb.jpg")
	err = cmd.Run()
	if err != nil {
		log.Print(err)
	}
	defer os.Remove("/tmp/thumb.jpg")

	nameSlice := strings.Split(vidKey, "/")
	typeSlice := strings.Split(nameSlice[1], ".")
	s := []string{nameSlice[0], "-thumb/", typeSlice[0], ".jpg"}
	thumbKey := strings.Join(s, "")
	dat, _ := ioutil.ReadFile("/tmp/thumb.jpg")

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
