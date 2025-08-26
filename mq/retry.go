package mq

import (
	"log"
	"math"
	"time"
)

func processWithRetry(body []byte, handler func([]byte) error, workerID int) {
	maxAttempts := 5
	initialInterval := 1000 * time.Millisecond
	multiplier := 2.0
	maxInterval := 10000 * time.Millisecond

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := handler(body)
		if err == nil {
			return
		}
		log.Printf("Worker %d: Attempt %d/%d failed: %v", workerID, attempt, maxAttempts, err)

		if attempt < maxAttempts {
			delay := float64(initialInterval) * math.Pow(multiplier, float64(attempt-1))
			if delay > float64(maxInterval) {
				delay = float64(maxInterval)
			}
			time.Sleep(time.Duration(delay))
		}
	}
	log.Printf("Worker %d: Message failed after %d attempts", workerID, maxAttempts)
}
