package messagequeue

import (
	"time"

	log "github.com/besanh/logger/logging/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type (
	INatsJetstream interface {
		Connect() error
		Ping()
	}

	NatsJetStream struct {
		NC     *nats.Conn
		Client *jetstream.JetStream
		Config Config
	}

	Config struct {
		Host string
	}
)

var NatJetstream INatsJetstream

func NewNatsJetstream(config Config) INatsJetstream {
	nat := &NatsJetStream{}

	return nat
}

func (n *NatsJetStream) Connect() error {
	nc, err := nats.Connect(n.Config.Host)
	if err != nil {
		return err
	}
	n.NC = nc
	n.Ping()
	return nil
}

func (n *NatsJetStream) Ping() {
	if err := nats.PingInterval(5 * time.Second); err != nil {
		log.Error(err)
	}
	log.Info("Ping nats success")
}
