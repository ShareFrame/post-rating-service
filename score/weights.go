package score

import (
	"fmt"
	"os"
	"strconv"
)

func MustEnvFloat(key string) float64 {
	v := os.Getenv(key)
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid float for %s: %v", key, err))
	}
	return f
}

func SafeIntWithBaseline(p *int, baseline int) int {
	if p == nil {
		return baseline
	}
	return *p
}

func SafeFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
