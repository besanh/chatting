package messagequeue

import (
	"context"
	"time"

	log "github.com/besanh/logger/logging/slog"
	"github.com/nats-io/nats.go"
)

type (
	INatsJetstream interface {
		Connect() error
		Ping()
		Publish(subject string, payload []byte) error
		Subscribe(ctx context.Context, subject, durable string, handler nats.MsgHandler) (*nats.Subscription, error)
	}

	NatsJetStream struct {
		Cfg Config
		Nc  *nats.Conn
		Js  nats.JetStreamContext
	}

	Config struct {
		Host           string
		PublishTimeout time.Duration
	}
)

// NewNatsJetstream constructs your JetStream client
func NewNatsJetstream(cfg Config) INatsJetstream {
	return &NatsJetStream{Cfg: cfg}
}

func (n *NatsJetStream) Connect() error {
	nc, err := nats.Connect(n.Cfg.Host)
	if err != nil {
		return err
	}
	n.Nc = nc

	js, err := nc.JetStream()
	if err != nil {
		return err
	}
	n.Js = js

	n.Ping()
	return nil
}

func (n *NatsJetStream) Ping() {
	if err := n.Nc.Flush(); err != nil {
		log.Error("nats flush failed:", err)
	} else {
		log.Info("nats flush (ping) succeeded")
	}
}

// Publish sends a message to the given subject via JetStream,
// using your configured timeout.
func (n *NatsJetStream) Publish(subject string, payload []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), n.Cfg.PublishTimeout)
	defer cancel()

	// this returns an Ack which you can inspect if needed
	if _, err := n.Js.Publish(subject, payload, nats.Context(ctx)); err != nil {
		return err
	}
	return nil
}

// Subscribe sets up a durable, manually-acked consumer on subject.
// handler will be called for each incoming message.
func (n *NatsJetStream) Subscribe(ctx context.Context, subject, durable string, handler nats.MsgHandler) (*nats.Subscription, error) {
	sub, err := n.Js.Subscribe(subject, handler, nats.Durable(durable), nats.ManualAck())
	if err != nil {
		return nil, err
	}

	// automatically unsubscribe when the context is done
	go func() {
		<-ctx.Done()
		sub.Unsubscribe()
	}()
	return sub, nil
}
