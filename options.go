package limiter

import (
	"time"

	"github.com/KlyuchnikovV/limiter/types/log"
)

type Option func(*Limiter) error

// WithCapacity - sets refillRate option to custom value
//
//	capacity should be greater that 0 (ErrCapacityIsLessThanOne will returned in other way)
func WithCapacity(capacity int64) Option {
	return func(l *Limiter) error {
		if capacity < 1 {
			return ErrCapacityIsLessThanOne
		}

		l.capacity = capacity

		return nil
	}
}

// WithRefillRate - sets refillRate option to custom value
//
//	refillRate should be greater that 0 (ErrRefillRateIsLessThanOne will returned in other way)
func WithRefillRate(rate time.Duration) Option {
	return func(l *Limiter) error {
		if rate < 1 {
			return ErrRefillRateIsLessThanOne
		}

		l.refillRate = rate

		return nil
	}
}

// WithLogger - sets refillRate option to custom value
//
//	log should not be nil (ErrProvidedLoggerIsNil will returned in other way)
func WithLogger(log log.Logger) Option {
	return func(l *Limiter) error {
		if log == nil {
			return ErrProvidedLoggerIsNil
		}

		l.log = log

		return nil
	}
}
