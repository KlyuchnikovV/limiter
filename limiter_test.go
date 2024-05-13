package limiter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KlyuchnikovV/limiter"
)

func assertError(t *testing.T, err, target error) {
	if !errors.Is(err, target) {
		t.Fatalf("errors are not equal: got: %s, expected: %s", err, target)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestSimpleLimiter(t *testing.T) {
	var id = "id"

	l, err := limiter.New(
		limiter.WithRefillRate(500*time.Millisecond),
		limiter.WithCapacity(3),
	)
	assertNoError(t, err)

	l.Start(context.Background())
	defer l.Stop()

	for i := 0; i < 3; i++ {
		_, err := l.Token(id)
		assertNoError(t, err)
	}

	_, err = l.Token(id)
	assertError(t, err, limiter.ErrTooManyRequests)

	time.Sleep(time.Second)

	_, err = l.Token(id)
	assertNoError(t, err)
}

func TestLimiterNotStarted(t *testing.T) {
	var id = "id"

	l, err := limiter.New()
	assertNoError(t, err)

	_, err = l.Token(id)
	assertError(t, err, limiter.ErrLimiterIsNotStarted)
}

func TestLimiterArguments(t *testing.T) {
	testCases := []struct {
		desc string

		refillRate time.Duration
		capacity   int64
		err        error
	}{
		{
			desc:       "Wrong refill rate",
			refillRate: 0,
			capacity:   1,
			err:        limiter.ErrRefillRateIsLessThanOne,
		},
		{
			desc:       "Wrong capacity",
			refillRate: 1,
			capacity:   0,
			err:        limiter.ErrCapacityIsLessThanOne,
		},
		{
			desc:       "All good",
			refillRate: 1,
			capacity:   1,
			err:        nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			_, err := limiter.New(
				limiter.WithRefillRate(tC.refillRate),
				limiter.WithCapacity(tC.capacity),
			)
			assertError(t, err, tC.err)
		})
	}
}
