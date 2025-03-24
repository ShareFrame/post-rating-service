package score

import (
	"log"
	"math"
	"time"

	"github.com/ShareFrame/post-rating-service/models"
)

var weights = map[string]float64{
	"watchTime":          MustEnvFloat("WEIGHT_WATCH_TIME"),
	"rewatches":          MustEnvFloat("WEIGHT_REWATCHES"),
	"shares":             MustEnvFloat("WEIGHT_SHARES"),
	"comments":           MustEnvFloat("WEIGHT_COMMENTS"),
	"likes":              MustEnvFloat("WEIGHT_LIKES"),
	"saves":              MustEnvFloat("WEIGHT_SAVES"),
	"engagementVelocity": MustEnvFloat("WEIGHT_ENGAGEMENT_VELOCITY"),
	"locationRelevance":  MustEnvFloat("WEIGHT_LOCATION_RELEVANCE"),
	"timeDecay":          MustEnvFloat("WEIGHT_TIME_DECAY"),
}

func CalculateTrendingScore(p models.Post) float64 {
	createdAt, err := time.Parse(time.RFC3339, p.CreatedAt)
	if err != nil {
		log.Printf("Invalid createdAt for post %s: %v", p.TID, err)
		return 0
	}
	hoursSince := time.Since(createdAt).Hours()

	baseScore := MustEnvFloat("BASELINE_SCORE")
	score := baseScore

	score += float64(SafeIntWithBaseline(p.WatchTime, 1)) * weights["watchTime"]
	score += float64(SafeIntWithBaseline(p.Rewatches, 0)) * weights["rewatches"]
	score += float64(SafeIntWithBaseline(p.Shares, 0)) * weights["shares"]
	score += float64(SafeIntWithBaseline(p.Comments, 0)) * weights["comments"]
	score += float64(SafeIntWithBaseline(p.Likes, 0)) * weights["likes"]
	score += float64(SafeIntWithBaseline(p.Saves, 0)) * weights["saves"]
	score += SafeFloat(p.EngagementVelocity) * weights["engagementVelocity"]
	score -= weights["timeDecay"] * math.Log(1+hoursSince)
	score += SafeFloat(p.LocationRelevance) * weights["locationRelevance"]

	totalMedia := len(p.ImageURIs) + len(p.VideoURIs)
	if totalMedia > 10 {
		unitPenalty := MustEnvFloat("MEDIA_OVERLOAD_UNIT_PENALTY")
		extra := totalMedia - 10
		penalty := float64(extra) * unitPenalty
		score -= penalty
		log.Printf("Applied media overload penalty to post %s (media=%d, penalty=%.2f)", p.TID, totalMedia, penalty)
	}

	return math.Round(score*10) / 10
}
