package service

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type (
	Subscriber struct {
		Address     string      `json:"address"`
		Message     chan []byte `json:"-"`
		CloseSlow   func()      `json:"-"`
		SubscribeAt time.Time   `json:"subscribe_at"`
	}

	Subscribers struct {
		SubscriberMessageBuffer int
		PublishLimiter          *rate.Limiter
		SubscribersMu           sync.Mutex
		Subscribers             SubscriberItem
	}

	SubscriberItem map[*Subscriber]struct{}
)

const (
	OAUTH2_TOKEN string = "oauth2_token"

	// State in callback url
	OAUTH2_STATE string = "chatting_state"
)

var (
	ENABLE_LOGIN_MULTI_SESSION bool = false

	API_SERVICE_NAME string = ""
	API_VERSION      string = ""

	// Wss
	ORIGIN_LIST = []string{"localhost:*"}
)
