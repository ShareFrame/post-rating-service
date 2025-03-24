package dynamo

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func ConvertStreamImage(av map[string]events.DynamoDBAttributeValue) (map[string]types.AttributeValue, error) {
	out := make(map[string]types.AttributeValue)
	for k, v := range av {
		converted, err := convertAttributeValue(v)
		if err != nil {
			return nil, fmt.Errorf("error converting key %s: %w", k, err)
		}
		out[k] = converted
	}
	return out, nil
}

func convertAttributeValue(v events.DynamoDBAttributeValue) (types.AttributeValue, error) {
	switch v.DataType() {
	case events.DataTypeString:
		return &types.AttributeValueMemberS{Value: v.String()}, nil
	case events.DataTypeNumber:
		return &types.AttributeValueMemberN{Value: v.Number()}, nil
	case events.DataTypeBoolean:
		return &types.AttributeValueMemberBOOL{Value: v.Boolean()}, nil
	case events.DataTypeNull:
		return &types.AttributeValueMemberNULL{Value: true}, nil
	case events.DataTypeList:
		list := v.List()
		convertedList := make([]types.AttributeValue, len(list))
		for i, item := range list {
			convertedItem, err := convertAttributeValue(item)
			if err != nil {
				return nil, err
			}
			convertedList[i] = convertedItem
		}
		return &types.AttributeValueMemberL{Value: convertedList}, nil
	case events.DataTypeMap:
		nested := v.Map()
		convertedMap := make(map[string]types.AttributeValue)
		for nk, nv := range nested {
			cv, err := convertAttributeValue(nv)
			if err != nil {
				return nil, err
			}
			convertedMap[nk] = cv
		}
		return &types.AttributeValueMemberM{Value: convertedMap}, nil
	default:
		return nil, fmt.Errorf("unsupported data type: %v", v.DataType())
	}
}
