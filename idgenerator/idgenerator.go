package idgenerator

import (
	"fmt"
	"math/rand"
	"time"
)

func Get(count int) []string {
	rand.Seed(time.Now().UnixNano())
	n := (rand.Uint64() >> 20) << 20
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = fmt.Sprintf("%X", n+uint64(i))
	}
	return result
}
