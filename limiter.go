package limiter

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/KlyuchnikovV/limiter/types/log"
)

var (
	ErrTooManyRequests         = fmt.Errorf("too many requests")
	ErrRefillRateIsLessThanOne = fmt.Errorf("refill rate is less than 1")
	ErrCapacityIsLessThanOne   = fmt.Errorf("refill rate is less than 1")
	ErrLimiterIsNotStarted     = fmt.Errorf("limiter is not started")
	ErrProvidedLoggerIsNil     = fmt.Errorf("provided logger is nil")

	defaultRefillRate = time.Second
	defaultCapacity   = 10
)

// Limiter - implementation of Rate Limiter
type Limiter struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	log log.Logger

	refillRate time.Duration
	capacity   int64

	numberOfRequests sync.Map
}

// New - creates new instance of Limiter
func New(options ...Option) (*Limiter, error) {
	var limiter = &Limiter{
		mu:               sync.Mutex{},
		log:              log.WrapSLog(slog.Default()).With("service", "limiter"),
		refillRate:       defaultRefillRate,
		capacity:         int64(defaultCapacity),
		numberOfRequests: sync.Map{},
	}

	for _, option := range options {
		if err := option(limiter); err != nil {
			return nil, fmt.Errorf("can't create new limiter: %w", err)
		}
	}

	return limiter, nil
}

// Start - starts refilling routine.
func (limiter *Limiter) Start(ctx context.Context) {
	if limiter.cancel != nil {
		limiter.log.Error("service was already started")

		return
	}

	limiter.ctx, limiter.cancel = context.WithCancel(ctx)

	go limiter.refill(limiter.log.With("routine", "refill"))

	limiter.log.Debug("service was started")
}

// Stop - stops service.
func (limiter *Limiter) Stop() {
	if limiter.cancel == nil {
		limiter.log.Error("service was already stopped")

		return
	}

	limiter.cancel()
	limiter.cancel = nil

	limiter.log.Debug("service was stopped")
}

// Token - returns token by request and increasing counter of used by id tokens
func (limiter *Limiter) Token(id string) (string, error) {
	if limiter.cancel == nil {
		return "", ErrLimiterIsNotStarted
	}

	limiter.log.Debug("trying to get token for object", "id", id)

	var numberOfRequests int64
	value, ok := limiter.numberOfRequests.Load(id)
	if ok {
		numberOfRequests = value.(int64)
	}

	if numberOfRequests >= limiter.capacity {
		limiter.log.Error("too many requests for object", "id", id)

		return "", ErrTooManyRequests
	}

	limiter.numberOfRequests.Store(id, numberOfRequests+1)

	limiter.log.Debug("token generated for object", "id", id)

	return limiter.generateToken(id), nil
}

func (limiter *Limiter) generateToken(id string) string {
	var sum = sha256.Sum256([]byte(
		fmt.Sprintf("%s:%d", id, time.Now().UnixNano()),
	))

	return string(sum[:])
}

func (limiter *Limiter) refill(log log.Logger) {
	var (
		shouldRun = true
		ticker    = time.NewTicker(limiter.refillRate)
	)

	defer ticker.Stop()

	log.Info("routine started")

	for shouldRun {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				log.Error("ticker refill routine was closed unexpectedly")
				shouldRun = false
				break
			}

			limiter.numberOfRequests.Range(func(id, value any) bool {
				log.Debug("started decreasing number of requests",
					"id", id,
				)

				var numberOfRequests = value.(int64)
				if numberOfRequests != 0 {
					log.Debug("number of requests decreased",
						"id", id,
						"numberOfRequests", numberOfRequests-1,
					)

					limiter.numberOfRequests.Store(id, numberOfRequests-1)
				} else {
					log.Debug("number of requests is zero",
						"id", id,
					)
				}

				return true
			})
		case <-limiter.ctx.Done():
			shouldRun = false
			break
		}
	}

	log.Info("refill routine stopped")
}
