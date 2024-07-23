package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"net/url"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/disintegration/imaging"
)

// Event is the input event structure
type Event struct {
	S3Bucket string `json:"s3Bucket"`
	S3Key    string `json:"s3Key"`
}

// Thumbnail is the output structure
type Thumbnail struct {
	S3Key    string `json:"s3key"`
	S3Bucket string `json:"s3bucket"`
}

// Response is the response structure
type Response struct {
	Thumbnail Thumbnail `json:"thumbnail"`
}

var (
	maxWidth        int
	maxHeight       int
	thumbnailBucket string
	s3Client        *s3.S3
)

func init() {
	// Initialize AWS session and S3 client
	sess := session.Must(session.NewSession())
	s3Client = s3.New(sess)

	// Read environment variables
	var err error
	maxWidth, err = strconv.Atoi(os.Getenv("MAX_WIDTH"))
	if err != nil {
		maxWidth = 250
	}

	maxHeight, err = strconv.Atoi(os.Getenv("MAX_HEIGHT"))
	if err != nil {
		maxHeight = 250
	}

	thumbnailBucket = os.Getenv("THUMBNAIL_BUCKET")
}

func handler(ctx context.Context, event Event) (Response, error) {
	// Decode S3 object key
	srcKey, err := url.QueryUnescape(event.S3Key)
	if err != nil {
		return Response{}, fmt.Errorf("failed to decode S3 key: %v", err)
	}

	// Get the object from S3
	getObjectOutput, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(event.S3Bucket),
		Key:    aws.String(srcKey),
	})
	if err != nil {
		return Response{}, fmt.Errorf("failed to get object from S3: %v", err)
	}
	defer getObjectOutput.Body.Close()

	// Decode the image
	img, _, err := image.Decode(getObjectOutput.Body)
	if err != nil {
		return Response{}, fmt.Errorf("failed to decode image: %v", err)
	}

	// Resize the image
	resizedImg := imaging.Resize(img, maxWidth, maxHeight, imaging.Lanczos)

	// Encode the resized image to a buffer
	var buf bytes.Buffer
	err = imaging.Encode(&buf, resizedImg, imaging.JPEG)
	if err != nil {
		return Response{}, fmt.Errorf("failed to encode resized image: %v", err)
	}

	// Upload the resized image to S3
	destKey := "resized-" + srcKey
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(thumbnailBucket),
		Key:         aws.String(destKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return Response{}, fmt.Errorf("failed to upload resized image to S3: %v", err)
	}

	// Return the response
	return Response{
		Thumbnail: Thumbnail{
			S3Key:    destKey,
			S3Bucket: thumbnailBucket,
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
