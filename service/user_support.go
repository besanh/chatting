package service

import (
	"errors"
	"time"

	"github.com/besanh/chatting/model"
	circuitbreaker "github.com/besanh/chatting/pkg/circuit_breaker"
	"github.com/sony/gobreaker/v2"
	"resty.dev/v3"
)

type (
	OAuth2Request struct {
		AccessToken string
		Url         string
		CBSetting   circuitbreaker.CBSetting
		Timeout     time.Duration

		// Device code
		ClientId   string
		Scope      string
		DeviceCode string
		Interval   int
		ExpiresIn  int
	}
)

/*
 * Combine Circuit Breaker pattern and resty
 */
func (s *User) getProfileUser(request OAuth2Request) (result model.UserProfile, err error) {
	client := resty.New()
	client.SetTimeout(request.Timeout)
	defer client.Close()

	cbSetting := circuitbreaker.CBGeneric(request.CBSetting)
	cb := gobreaker.NewCircuitBreaker[model.UserProfile](*cbSetting)

	result, err = cb.Execute(func() (res model.UserProfile, err error) {
		resp, err := client.R().
			SetHeaders(map[string]string{
				"Authorization": "Bearer " + request.AccessToken,
				"Content-Type":  "application/json",
				"Accept":        "application/json",
			}).
			SetResult(&res).
			Get(request.Url)
		if err != nil {
			return
		} else if resp.IsError() || resp.StatusCode() != 200 {
			err = errors.New(resp.Status())
			return
		}

		return
	})

	if err != nil {
		return
	}

	return
}
