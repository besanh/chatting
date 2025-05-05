package circuitbreaker

import (
	"time"

	"github.com/sony/gobreaker/v2"
)

type (
	CBSetting struct {
		CBName        string        `json:"cb_name"`
		MaxRequest    uint32        `json:"max_request"`
		Interval      time.Duration `json:"interval"`
		TimeOut       time.Duration `json:"timeout"`
		MaxTripCB     int           `json:"max_trip_cb"`
		OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
		IsSuccessful  func(err error) bool
	}
)

func CBGeneric(setting CBSetting) *gobreaker.Settings {
	return &gobreaker.Settings{
		Name:        setting.CBName,
		MaxRequests: setting.MaxRequest,
		Interval:    setting.Interval,
		Timeout:     setting.TimeOut,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip the breaker if there are at least max_trip_cb requests and failure ratio is 50% or more.
			if counts.Requests < uint32(setting.MaxTripCB) {
				return false
			}

			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= 0.5
		},
		OnStateChange: setting.OnStateChange,
		IsSuccessful:  setting.IsSuccessful,
	}
}
