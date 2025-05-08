package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/besanh/chatting/model"
	circuitbreaker "github.com/besanh/chatting/pkg/circuit_breaker"
	"resty.dev/v3"
)

type (
	OAuth2Request struct {
		AccessToken  string
		RefreshToken string
		Url          string
		CBSetting    circuitbreaker.CBSetting
		Timeout      time.Duration

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

	cb := circuitbreaker.NewCB(request.CBSetting)

	res, err := cb.Execute(func() (err error) {
		resp, err := client.R().
			SetHeaders(map[string]string{
				"Authorization": "Bearer " + request.AccessToken,
				"Content-Type":  "application/json",
				"Accept":        "application/json",
			}).
			// SetResult(&res).
			Get(request.Url)
		if err != nil {
			return
		} else if resp.IsError() || resp.StatusCode() != 200 {
			err = errors.New(resp.Status())
			return
		}

		return
	})

	if res != nil {
		result = res.(model.UserProfile)
	}

	if err != nil {
		return
	}

	return
}

func (s *User) revokeGoogleToken(request OAuth2Request) error {
	client := resty.New()
	client.SetTimeout(request.Timeout)
	defer client.Close()

	// Wrap the HTTP call in your circuit-breaker
	cb := circuitbreaker.NewCB(request.CBSetting)

	_, err := cb.Execute(func() (err error) {
		resp, err := client.R().
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetFormData(map[string]string{
				"token": request.RefreshToken,
			}).
			Post(request.Url)
		if err != nil {
			return
		}
		if resp.StatusCode() != http.StatusOK {
			err = fmt.Errorf("oauth2 revoke failed: %s", resp.Status())
			return
		}
		return
	})

	return err
}
