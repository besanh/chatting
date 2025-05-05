package model

type (
	OAuth2Callback struct {
		Code  string `json:"code"`
		State string `json:"state"`
		Scope string `json:"scope"`
	}

	MiddlwareOauth2 struct {
		Authenticated bool   `json:"authenticated"`
		UserID        string `json:"user_id"`
		DeviceID      string `json:"device_id"`
	}
)
