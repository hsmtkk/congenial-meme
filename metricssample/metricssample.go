package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func main() {
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	client := cloudwatch.NewFromConfig(sdkConfig)
	dimension := types.Dimension{
		Name:  aws.String("url"),
		Value: aws.String("https://www.example.com"),
	}
	metric := types.MetricDatum{
		MetricName: aws.String("rtt"),
		Dimensions: []types.Dimension{dimension},
		Unit:       types.StandardUnitMilliseconds,
		Value:      aws.Float64(rand.Float64()),
	}
	result, err := client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String("url_rtt"),
		MetricData: []types.MetricDatum{metric},
	})
	if err != nil {
		log.Fatalf("failed to put metric data: %v", err)
	}
	fmt.Printf("result: %v\n", result.ResultMetadata)
}
