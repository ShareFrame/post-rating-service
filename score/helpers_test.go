package score_test

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/ShareFrame/post-rating-service/models"
	"github.com/ShareFrame/post-rating-service/score"
	"github.com/stretchr/testify/assert"
)

func TestSafeIntWithBaseline(t *testing.T) {
	tests := []struct {
		name     string
		input    *int
		baseline int
		expected int
	}{
		{"nil pointer", nil, 7, 7},
		{"non-nil pointer", ptrInt(3), 10, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := score.SafeIntWithBaseline(tt.input, tt.baseline)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    *float64
		expected float64
	}{
		{"nil float", nil, 0},
		{"non-nil float", ptrFloat(2.5), 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := score.SafeFloat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMustEnvFloat(t *testing.T) {
	const key = "TEST_FLOAT_ENV"

	t.Run("valid float", func(t *testing.T) {
		os.Setenv(key, "1.23")
		defer os.Unsetenv(key)

		result := score.MustEnvFloat(key)
		assert.Equal(t, 1.23, result)
	})

	t.Run("invalid float", func(t *testing.T) {
		os.Setenv(key, "not-a-float")
		defer os.Unsetenv(key)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for invalid float, but didn't get one")
			}
		}()

		score.MustEnvFloat(key)
	})

	t.Run("unset env var", func(t *testing.T) {
		os.Unsetenv(key)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for missing env var, but didn't get one")
			}
		}()

		score.MustEnvFloat(key)
	})
}

func TestCalculateTrendingScore(t *testing.T) {
	now := time.Date(2025, 3, 23, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		post     models.Post
		weights  map[string]float64
		expected float64
	}{
		{
			name: "baseline only",
			post: models.Post{
				TID:       "p1",
				CreatedAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
			},
			weights: map[string]float64{
				"BASELINE_SCORE":              5,
				"WEIGHT_WATCH_TIME":           0,
				"WEIGHT_REWATCHES":            0,
				"WEIGHT_SHARES":               0,
				"WEIGHT_COMMENTS":             0,
				"WEIGHT_LIKES":                0,
				"WEIGHT_SAVES":                0,
				"WEIGHT_ENGAGEMENT_VELOCITY":  0,
				"WEIGHT_LOCATION_RELEVANCE":   0,
				"WEIGHT_TIME_DECAY":           0,
				"MEDIA_OVERLOAD_UNIT_PENALTY": 0,
			},
			expected: 5,
		},
		{
			name: "engagement and time decay",
			post: models.Post{
				TID:       "p2",
				CreatedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
				Likes:     ptrInt(10),
				Comments:  ptrInt(2),
				Shares:    ptrInt(1),
			},
			weights: map[string]float64{
				"BASELINE_SCORE":              0,
				"WEIGHT_WATCH_TIME":           0,
				"WEIGHT_REWATCHES":            0,
				"WEIGHT_SHARES":               1,
				"WEIGHT_COMMENTS":             0.5,
				"WEIGHT_LIKES":                0.2,
				"WEIGHT_SAVES":                0,
				"WEIGHT_ENGAGEMENT_VELOCITY":  0,
				"WEIGHT_LOCATION_RELEVANCE":   0,
				"WEIGHT_TIME_DECAY":           1,
				"MEDIA_OVERLOAD_UNIT_PENALTY": 0,
			},
			expected: math.Round((10*0.2+2*0.5+1*1-1*math.Log(1+2))*10) / 10,
		},
		{
			name: "media overload penalty",
			post: models.Post{
				TID:       "p3",
				CreatedAt: now.Format(time.RFC3339),
				ImageURIs: make([]string, 8),
				VideoURIs: make([]string, 5),
				Likes:     ptrInt(4),
				WatchTime: ptrInt(30),
			},
			weights: map[string]float64{
				"BASELINE_SCORE":              2,
				"WEIGHT_WATCH_TIME":           0.1,
				"WEIGHT_REWATCHES":            0,
				"WEIGHT_SHARES":               0,
				"WEIGHT_COMMENTS":             0,
				"WEIGHT_LIKES":                0.5,
				"WEIGHT_SAVES":                0,
				"WEIGHT_ENGAGEMENT_VELOCITY":  0,
				"WEIGHT_LOCATION_RELEVANCE":   0,
				"WEIGHT_TIME_DECAY":           0,
				"MEDIA_OVERLOAD_UNIT_PENALTY": 0.25,
			},
			expected: math.Round((2+30*0.1+4*0.5-(3*0.25))*10) / 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score.InjectWeights(tt.weights)
			actual := score.CalculateTrendingScore(tt.post, now)
			assert.InEpsilon(t, tt.expected, actual, 0.01, "Expected %.2f, got %.2f", tt.expected, actual)
		})
	}
}

func ptrInt(v int) *int           { return &v }
func ptrFloat(v float64) *float64 { return &v }
