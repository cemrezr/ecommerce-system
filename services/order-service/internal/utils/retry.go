package utils

import (
	"time"

	"github.com/sony/gobreaker"
)

func RetryWithBreaker(cb *gobreaker.CircuitBreaker, op func() error) (int, error) {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, op()
		})
		if err == nil {
			return i + 1, nil
		}
		lastErr = err
		time.Sleep(retryDelay)
	}
	return maxRetries, lastErr
}
