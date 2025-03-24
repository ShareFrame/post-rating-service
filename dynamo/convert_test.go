package dynamo_test

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ShareFrame/post-rating-service/dynamo"
	"github.com/stretchr/testify/assert"
)

func TestConvertStreamImage(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]events.DynamoDBAttributeValue
		expected map[string]types.AttributeValue
		wantErr  bool
	}{
		{
			name: "simple string",
			input: map[string]events.DynamoDBAttributeValue{
				"tid": events.NewStringAttribute("abc123"),
			},
			expected: map[string]types.AttributeValue{
				"tid": &types.AttributeValueMemberS{Value: "abc123"},
			},
			wantErr: false,
		},
		{
			name: "number + boolean + null",
			input: map[string]events.DynamoDBAttributeValue{
				"likes":   events.NewNumberAttribute("10"),
				"active":  events.NewBooleanAttribute(true),
				"deleted": events.NewNullAttribute(),
			},
			expected: map[string]types.AttributeValue{
				"likes":   &types.AttributeValueMemberN{Value: "10"},
				"active":  &types.AttributeValueMemberBOOL{Value: true},
				"deleted": &types.AttributeValueMemberNULL{Value: true},
			},
			wantErr: false,
		},
		{
			name: "list of numbers",
			input: map[string]events.DynamoDBAttributeValue{
				"values": events.NewListAttribute(
					[]events.DynamoDBAttributeValue{
						events.NewNumberAttribute("1"),
						events.NewNumberAttribute("2"),
					}),
			},
			expected: map[string]types.AttributeValue{
				"values": &types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberN{Value: "1"},
					&types.AttributeValueMemberN{Value: "2"},
				}},
			},
			wantErr: false,
		},
		{
			name: "nested map",
			input: map[string]events.DynamoDBAttributeValue{
				"meta": events.NewMapAttribute(map[string]events.DynamoDBAttributeValue{
					"version": events.NewStringAttribute("v1"),
				}),
			},
			expected: map[string]types.AttributeValue{
				"meta": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
					"version": &types.AttributeValueMemberS{Value: "v1"},
				}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dynamo.ConvertStreamImage(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
