package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	s3Client := s3.NewFromConfig(sdkConfig)
	result, err := s3Client.ListBuckets(ctx, nil)
	if err != nil {
		log.Fatalf("failed to list buckets: %v", err)
	}
	for _, bucket := range result.Buckets {
		fmt.Printf("bucket: %s\n", *bucket.Name)
	}
}
