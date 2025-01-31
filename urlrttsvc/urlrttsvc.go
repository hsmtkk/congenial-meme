package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s config.yaml", os.Args[0])
	}
	confYaml := os.Args[1]
	conf, err := readConfig(confYaml)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	client := cloudwatch.NewFromConfig(sdkConfig)

	for {
		for _, url := range conf.URLs {
			rtt, err := poll(url)
			if err != nil {
				log.Printf("failed to poll %s: %v", url, err)
				continue
			}
			if err := pushMetrics(ctx, client, url, rtt); err != nil {
				log.Printf("failed to push metrics for %s: %v", url, err)
				continue
			}
			log.Printf("pushed metrics for %s: %d milliseconds", url, rtt)
		}
		time.Sleep(time.Duration(conf.PollIntervalSecond) * time.Second)
	}
}

type serviceConfig struct {
	PollIntervalSecond int      `yaml:"poll_interval_second"`
	URLs               []string `yaml:"urls"`
}

func readConfig(confYaml string) (serviceConfig, error) {
	f, err := os.Open(confYaml)
	if err != nil {
		return serviceConfig{}, fmt.Errorf("failed to open config file %s: %w", confYaml, err)
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	conf := serviceConfig{}
	if err := decoder.Decode(&conf); err != nil {
		return serviceConfig{}, fmt.Errorf("failed to decode config file %s: %w", confYaml, err)
	}
	return conf, nil
}

func poll(url string) (int64, error) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get %s: %w", url, err)
	}
	defer resp.Body.Close()
	if _, err := io.ReadAll(resp.Body); err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}
	elapsed := time.Since(start).Milliseconds()
	return elapsed, nil
}

func pushMetrics(ctx context.Context, client *cloudwatch.Client, url string, rttMilliseconds int64) error {
	dimension := types.Dimension{
		Name:  aws.String("url"),
		Value: aws.String(url),
	}
	metric := types.MetricDatum{
		MetricName: aws.String("rtt"),
		Dimensions: []types.Dimension{dimension},
		Unit:       types.StandardUnitMilliseconds,
		Value:      aws.Float64(float64(rttMilliseconds)),
	}
	_, err := client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String("url_rtt"),
		MetricData: []types.MetricDatum{metric},
	})
	if err != nil {
		return fmt.Errorf("failed to push metric data: %w", err)
	}
	return nil
}
