package score

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/ShareFrame/post-rating-service/models"
)

var weights = make(map[string]float64)

func InjectWeights(custom map[string]float64) {
	weights = custom
}

func LoadWeightsFromEnv() {
	keys := []string{
		"WEIGHT_WATCH_TIME", "WEIGHT_REWATCHES", "WEIGHT_SHARES", "WEIGHT_COMMENTS",
		"WEIGHT_LIKES", "WEIGHT_SAVES", "WEIGHT_ENGAGEMENT_VELOCITY", "WEIGHT_LOCATION_RELEVANCE",
		"WEIGHT_TIME_DECAY", "MEDIA_OVERLOAD_UNIT_PENALTY", "BASELINE_SCORE",
	}

	for _, key := range keys {
		weights[key] = mustEnvFloat(key)
	}
}

func mustEnvFloat(key string) float64 {
	v := os.Getenv(key)
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid float for %s: %v", key, err))
	}
	return f
}

func CalculateTrendingScore(p models.Post, now time.Time) float64 {
	createdAt, err := time.Parse(time.RFC3339, p.CreatedAt)
	if err != nil {
		return 0
	}
	hoursSince := now.Sub(createdAt).Hours()

	score := weights["BASELINE_SCORE"]
	score += float64(safeInt(p.WatchTime, 1)) * weights["WEIGHT_WATCH_TIME"]
	score += float64(safeInt(p.Rewatches, 0)) * weights["WEIGHT_REWATCHES"]
	score += float64(safeInt(p.Shares, 0)) * weights["WEIGHT_SHARES"]
	score += float64(safeInt(p.Comments, 0)) * weights["WEIGHT_COMMENTS"]
	score += float64(safeInt(p.Likes, 0)) * weights["WEIGHT_LIKES"]
	score += float64(safeInt(p.Saves, 0)) * weights["WEIGHT_SAVES"]
	score += safeFloat(p.EngagementVelocity) * weights["WEIGHT_ENGAGEMENT_VELOCITY"]
	score += safeFloat(p.LocationRelevance) * weights["WEIGHT_LOCATION_RELEVANCE"]
	score -= weights["WEIGHT_TIME_DECAY"] * math.Log(1+hoursSince)

	totalMedia := len(p.ImageURIs) + len(p.VideoURIs)
	if totalMedia > 10 {
		extra := totalMedia - 10
		score -= float64(extra) * weights["MEDIA_OVERLOAD_UNIT_PENALTY"]
	}

	return math.Round(score*10) / 10
}

func safeInt(p *int, baseline int) int {
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
