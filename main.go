package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Post struct {
	TID                string   `dynamodbav:"tid"`
	CreatedAt          string   `dynamodbav:"createdAt"`
	Likes              *int     `dynamodbav:"likes"`
	Shares             *int     `dynamodbav:"shares"`
	Comments           *int     `dynamodbav:"comments"`
	Rewatches          *int     `dynamodbav:"rewatches"`
	Saves              *int     `dynamodbav:"saves"`
	WatchTime          *int     `dynamodbav:"watchTime"`
	EngagementVelocity *float64 `dynamodbav:"engagementVelocity"`
	LocationRelevance  *float64 `dynamodbav:"locationRelevance"`
	TrendingScore      *float64 `dynamodbav:"trendingScore"`
	ImageURIs          []string `dynamodbav:"imageUris"`
	VideoURIs          []string `dynamodbav:"videoUris"`
}

var (
	dynamoClient *dynamodb.Client
	tableName    = os.Getenv("DYNAMODB_TABLE")
	weights      = map[string]float64{
		"watchTime":          mustEnvFloat("WEIGHT_WATCH_TIME"),
		"rewatches":          mustEnvFloat("WEIGHT_REWATCHES"),
		"shares":             mustEnvFloat("WEIGHT_SHARES"),
		"comments":           mustEnvFloat("WEIGHT_COMMENTS"),
		"likes":              mustEnvFloat("WEIGHT_LIKES"),
		"saves":              mustEnvFloat("WEIGHT_SAVES"),
		"engagementVelocity": mustEnvFloat("WEIGHT_ENGAGEMENT_VELOCITY"),
		"locationRelevance":  mustEnvFloat("WEIGHT_LOCATION_RELEVANCE"),
		"timeDecay":          mustEnvFloat("WEIGHT_TIME_DECAY"),
	}
)

func mustEnvFloat(key string) float64 {
	v := os.Getenv(key)
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		log.Fatalf("invalid float for %s: %v", key, err)
	}
	return f
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
}

func handler(ctx context.Context, event events.DynamoDBEvent) error {
	for _, record := range event.Records {
		if record.EventName != "INSERT" && record.EventName != "MODIFY" {
			continue
		}

		convertedImage, err := convertStreamImage(record.Change.NewImage)
		if err != nil {
			log.Printf("error converting NewImage: %v", err)
			continue
		}

		var post Post
		err = attributevalue.UnmarshalMap(convertedImage, &post)
		if err != nil {
			log.Printf("failed to unmarshal record: %v", err)
			dump, _ := json.MarshalIndent(convertedImage, "", "  ")
			log.Printf("convertedImage: %s", dump)
			continue
		}

		score := calculateTrendingScore(post)
		post.TrendingScore = &score

		log.Printf("Recalculated trendingScore for tid=%s: %.2f", post.TID, score)

		_, err = dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(tableName),
			Key: map[string]ddbtypes.AttributeValue{
				"tid": &ddbtypes.AttributeValueMemberS{Value: post.TID},
			},
			UpdateExpression: aws.String("SET trendingScore = :score"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":score": &ddbtypes.AttributeValueMemberN{Value: fmt.Sprintf("%.1f", score)},
			},
		})
		if err != nil {
			log.Printf("failed to update trending score: %v", err)
		}
	}
	return nil
}

func convertStreamImage(av map[string]events.DynamoDBAttributeValue) (map[string]ddbtypes.AttributeValue, error) {
	out := make(map[string]ddbtypes.AttributeValue)
	for k, v := range av {
		converted, err := convertAttributeValue(v)
		if err != nil {
			return nil, fmt.Errorf("error converting key %s: %w", k, err)
		}
		out[k] = converted
	}
	return out, nil
}

func convertAttributeValue(v events.DynamoDBAttributeValue) (ddbtypes.AttributeValue, error) {
	switch v.DataType() {
	case events.DataTypeString:
		return &ddbtypes.AttributeValueMemberS{Value: v.String()}, nil
	case events.DataTypeNumber:
		return &ddbtypes.AttributeValueMemberN{Value: v.Number()}, nil
	case events.DataTypeBoolean:
		return &ddbtypes.AttributeValueMemberBOOL{Value: v.Boolean()}, nil
	case events.DataTypeNull:
		return &ddbtypes.AttributeValueMemberNULL{Value: true}, nil
	case events.DataTypeList:
		list := v.List()
		convertedList := make([]ddbtypes.AttributeValue, len(list))
		for i, item := range list {
			convertedItem, err := convertAttributeValue(item)
			if err != nil {
				return nil, fmt.Errorf("error converting list item %d: %w", i, err)
			}
			convertedList[i] = convertedItem
		}
		return &ddbtypes.AttributeValueMemberL{Value: convertedList}, nil
	case events.DataTypeMap:
		nested := v.Map()
		convertedMap := make(map[string]ddbtypes.AttributeValue)
		for nk, nv := range nested {
			convertedVal, err := convertAttributeValue(nv)
			if err != nil {
				return nil, fmt.Errorf("error converting map key %s: %w", nk, err)
			}
			convertedMap[nk] = convertedVal
		}
		return &ddbtypes.AttributeValueMemberM{Value: convertedMap}, nil
	default:
		return nil, fmt.Errorf("unsupported data type: %s", v.DataType())
	}
}

func safeIntWithBaseline(p *int, baseline int) int {
	if p == nil {
		return baseline
	}
	return *p
}

func safeFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func calculateTrendingScore(p Post) float64 {
	createdAt, err := time.Parse(time.RFC3339, p.CreatedAt)
	if err != nil {
		log.Printf("Invalid createdAt for post %s: %v", p.TID, err)
		return 0
	}
	hoursSince := time.Since(createdAt).Hours()

	baseScore := mustEnvFloat("BASELINE_SCORE")
	score := baseScore

	score += float64(safeIntWithBaseline(p.WatchTime, 1)) * weights["watchTime"]
	score += float64(safeIntWithBaseline(p.Rewatches, 0)) * weights["rewatches"]
	score += float64(safeIntWithBaseline(p.Shares, 0)) * weights["shares"]
	score += float64(safeIntWithBaseline(p.Comments, 0)) * weights["comments"]
	score += float64(safeIntWithBaseline(p.Likes, 0)) * weights["likes"]
	score += float64(safeIntWithBaseline(p.Saves, 0)) * weights["saves"]
	score += safeFloat(p.EngagementVelocity) * weights["engagementVelocity"]
	score -= weights["timeDecay"] * math.Log(1+hoursSince)
	score += safeFloat(p.LocationRelevance) * weights["locationRelevance"]

	totalMedia := len(p.ImageURIs) + len(p.VideoURIs)
	if totalMedia > 10 {
		unitPenalty := mustEnvFloat("MEDIA_OVERLOAD_UNIT_PENALTY")
		extra := totalMedia - 10
		penalty := float64(extra) * unitPenalty
		score -= penalty
		log.Printf("Applied media overload penalty to post %s (media=%d, penalty=%.2f)", p.TID, totalMedia, penalty)
	}

	return math.Round(score*10) / 10
}

func main() {
	lambda.Start(handler)
}
