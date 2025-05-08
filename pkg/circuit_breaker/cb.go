package circuitbreaker

import (
	"time"

	"github.com/sony/gobreaker/v2"
)

// CBSetting holds configuration for the circuit breaker, including
// thresholds, timeouts, and callbacks.
type CBSetting struct {
	CBName        string        `json:"cb_name"`
	MaxRequest    uint32        `json:"max_request"`
	Interval      time.Duration `json:"interval"`
	TimeOut       time.Duration `json:"timeout"`
	MaxTripCB     int           `json:"max_trip_cb"`
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
	IsSuccessful  func(err error) bool
}

// CBGeneric converts CBSetting into gobreaker.Settings
func (s *CBSetting) CBGeneric() *gobreaker.Settings {
	return &gobreaker.Settings{
		Name:        s.CBName,
		MaxRequests: s.MaxRequest,
		Interval:    s.Interval,
		Timeout:     s.TimeOut,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < uint32(s.MaxTripCB) {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= 0.5
		},
		OnStateChange: s.OnStateChange,
		IsSuccessful:  s.IsSuccessful,
	}
}

// CB wraps a gobreaker.CircuitBreaker[any] and provides
// a simple Execute API for functions that return only an error.
type CB struct {
	breaker *gobreaker.CircuitBreaker[any]
}

// NewCB creates a new CB (circuit breaker) from settings.
func NewCB(setting CBSetting) *CB {
	// Instantiate generic circuit breaker with any type
	breaker := gobreaker.NewCircuitBreaker[any](*setting.CBGeneric())
	return &CB{breaker: breaker}
}

// Execute runs the given function under circuit breaker control.
// It returns fn's error or gobreaker.ErrOpenState if the circuit is open.
func (c *CB) Execute(fn func() error) (result any, err error) {
	result, err = c.breaker.Execute(func() (any, error) {
		return nil, fn()
	})
	return
}
