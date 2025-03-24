package dynamo

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func UpdateTrendingScore(client *dynamodb.Client, tableName, tid string, score float64) error {
	_, err := client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"tid": &types.AttributeValueMemberS{Value: tid},
		},
		UpdateExpression: aws.String("SET trendingScore = :score"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":score": &types.AttributeValueMemberN{Value: fmt.Sprintf("%.1f", score)},
		},
	})
	return err
}
