package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/besanh/chatting/common/caching"
	"github.com/besanh/chatting/common/util"
	"github.com/besanh/chatting/config"
	translate "github.com/besanh/chatting/external/google"
	"github.com/besanh/chatting/model"
	circuitbreaker "github.com/besanh/chatting/pkg/circuit_breaker"
	log "github.com/besanh/logger/logging/slog"
	"github.com/nats-io/nats.go"
)

type (
	ISubscriber interface {
		Subscribe(ctx context.Context, userID, sessionID string, ch chan<- *model.WsEvent) (err error)
		Publish(ctx context.Context, userID string, evt *model.WsEvent) (err error)
		HandleGoogleApi(request model.TranslationRequest) (resp *model.TranslationResponse, err error)
	}
	SubscriberService struct {
		google translate.IGoogleTranslate
		cfg    config.Config
		cb     *circuitbreaker.CB
	}
)

var (
	SubscriberServiceGlobal ISubscriber
	WsSubscribers           *Subscribers
	CHATTING_CONNECTION_KEY string = "chatting_connection"
)

func NewSubscriberService(cfg config.Config, cbSetting *circuitbreaker.CBSetting, google translate.IGoogleTranslate) ISubscriber {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	addressKey, err := caching.RCache.Keys(ctx, CHATTING_CONNECTION_KEY)
	if err != nil {
		log.Error(err)
		panic(err)
	} else if len(addressKey) > 0 {
		if err := caching.RCache.Del(ctx, addressKey); err != nil {
			log.Error(err)
			panic(err)
		}
	}

	return &SubscriberService{
		google: google,
		cfg:    cfg,
		cb:     circuitbreaker.NewCB(*cbSetting),
	}
}

// Subscribe creates a durable, manually-acked consumer on subject "user_events.<userID>",
// using sessionID as the durable name. Messages are JSON-unmarshaled into model.WsEvent
// and sent on ch. On ctx.Done() the subscription is torn down.
func (s *SubscriberService) Subscribe(ctx context.Context, userID, sessionID string, ch chan<- *model.WsEvent) (err error) {
	subject := fmt.Sprintf("user_events.%s", userID)
	durable := sessionID

	// wrap subscribe in circuit breaker
	_, err = s.cb.Execute(func() (res any, err error) {
		sub, err := (*s.cfg.Pkg.NatJetstream.Js).Subscribe(subject, func(msg *nats.Msg) {
			var evt model.WsEvent
			if err := json.Unmarshal(msg.Data, &evt); err != nil {
				log.Error(err)
				msg.Ack() // drop bad payload
				return
			}

			select {
			case ch <- &evt:
				msg.Ack()
			case <-ctx.Done():
				msg.Nak() // redeliver
			}
		},
			nats.Durable(durable),
			nats.ManualAck(),
		)
		if err != nil {
			err = fmt.Errorf("jetstream subscribe: %w", err)
			log.Error(err)
			return
		}
		// cleanup
		go func() {
			<-ctx.Done()
			sub.Unsubscribe()
		}()

		return
	})

	return
}

// Publish serializes evt and publishes it to "user_events.<userID>".
func (s *SubscriberService) Publish(ctx context.Context, userID string, evt *model.WsEvent) (err error) {
	payload, err := json.Marshal(evt)
	if err != nil {
		err = fmt.Errorf("marshal event: %w", err)
		log.Error(err)
		return
	}
	subject := fmt.Sprintf("user_events.%s", userID)

	_, err = s.cb.Execute(func() (res any, err error) {
		_, err = (*s.cfg.Pkg.NatJetstream.Js).Publish(subject, payload)
		log.Error(err)
		return

	})
	if err != nil {
		err = fmt.Errorf("jetstream publish: %w", err)
		log.Error(err)
		return
	}
	return
}

func (s *SubscriberService) HandleGoogleApi(request model.TranslationRequest) (resp *model.TranslationResponse, err error) {
	if len(s.cfg.Pkg.Translate.Url) > 0 {
		item := util.RandomChoice(s.cfg.Pkg.Translate.Url)
		method, clientKey, url, errTmp := parseMethodKeyURL(item)
		if errTmp != nil {
			err = errTmp
			log.Error(err)
			return
		}
		log.Info("method: ", method, ", clientKey: ", clientKey, ", url: ", url)

		request := model.GoogleTranslateApiRequest{
			Url:    url,
			Client: clientKey,
			Sl:     "auto",
			Tl:     request.LangCode,
			Dt:     "t",
			Q:      request.Content,
			Engine: "wt_lib",
		}

		if method == "get" {
			log.Infof("getgoogleApi: %v", url)
			resp, err = s.google.GetGoogleTranslateApi(request)
			if err != nil {
				log.Error(err)
				return
			}
		} else if method == "post" {
			log.Infof("postgoogleApi: %v", url)
			body := model.ToNestedJSON(request)
			resp, err = s.google.PostGoogleTranslateApi(url, clientKey, body)
			if err != nil {
				log.Error(err)
				return
			}
		} else {
			err = fmt.Errorf("invalid input format, expected method and URL separated by a comma")
			log.Error(err)
			return
		}
	} else {
		err = fmt.Errorf("google translate url is empty")
		log.Error(err)
	}

	return
}

func parseMethodKeyURL(raw string) (method, clientKey, urlStr string, err error) {
	parts := strings.SplitN(raw, ",", 3)
	if len(parts) != 3 {
		err = fmt.Errorf("invalid format %q, expected method,clientKey,URL", raw)
		log.Error(err)
		return
	}
	method = strings.ToLower(strings.TrimSpace(parts[0]))
	clientKey = strings.TrimSpace(parts[1])
	urlStr = strings.TrimSpace(parts[2])
	return
}
