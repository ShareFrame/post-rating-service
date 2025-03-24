package models

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
