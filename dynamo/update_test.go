package dynamo_test

import (
	"context"
	"testing"

	"github.com/ShareFrame/post-rating-service/dynamo"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

type mockDynamoClient struct {
	UpdateItemFn func(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

func (m *mockDynamoClient) UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return m.UpdateItemFn(ctx, input, optFns...)
}

func TestUpdateTrendingScore(t *testing.T) {
	mockClient := &mockDynamoClient{
		UpdateItemFn: func(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
			assert.Equal(t, "posts-table", *input.TableName)
			assert.Equal(t, &types.AttributeValueMemberS{Value: "abc123"}, input.Key["tid"])
			assert.Equal(t, "SET trendingScore = :score", *input.UpdateExpression)
			assert.Equal(t, &types.AttributeValueMemberN{Value: "12.3"}, input.ExpressionAttributeValues[":score"])
			return &dynamodb.UpdateItemOutput{}, nil
		},
	}

	err := dynamo.UpdateTrendingScore(mockClient, "posts-table", "abc123", 12.3)
	assert.NoError(t, err)
}
