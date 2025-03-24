package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/ShareFrame/post-rating-service/dynamo"
	"github.com/ShareFrame/post-rating-service/models"
	"github.com/ShareFrame/post-rating-service/score"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

var (
	client    *dynamodb.Client
	tableName = os.Getenv("DYNAMODB_TABLE")
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	client = dynamodb.NewFromConfig(cfg)
}

func handler(ctx context.Context, event events.DynamoDBEvent) error {
	for _, record := range event.Records {
		if record.EventName != "INSERT" && record.EventName != "MODIFY" {
			continue
		}

		image, err := dynamo.ConvertStreamImage(record.Change.NewImage)
		if err != nil {
			log.Printf("error converting image: %v", err)
			continue
		}

		var post models.Post
		if err := attributevalue.UnmarshalMap(image, &post); err != nil {
			log.Printf("unmarshal error: %v", err)
			data, _ := json.MarshalIndent(image, "", "  ")
			log.Printf("raw data: %s", data)
			continue
		}

		scoreValue := score.CalculateTrendingScore(post, time.Now())
		log.Printf("Score for %s = %.2f", post.TID, scoreValue)

		if err := dynamo.UpdateTrendingScore(client, tableName, post.TID, scoreValue); err != nil {
			log.Printf("update error for %s: %v", post.TID, err)
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
