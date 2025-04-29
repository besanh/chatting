package messagequeue

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/adjust/rmq/v5"
	"github.com/redis/go-redis/v9"
)

type Rcfg struct {
	Address  string
	Username string
	Password string
	DB       int
}

var RMQ *RMQConnection

const (
	tag           = "rmq"
	pollDuration  = 100 * time.Millisecond
	prefetchLimit = 1000
)

type RMQConnection struct {
	RedisClient *redis.Client
	Config      Rcfg
	Conn        rmq.Connection
	Queues      map[string]rmq.Queue
	Server      *RMQServer
	Client      *RMQClient
}

type RMQServer struct {
	mu     sync.Mutex
	conn   rmq.Connection
	Queues map[string]rmq.Queue
}

type RMQClient struct {
	mu     sync.Mutex // protects Queues
	conn   rmq.Connection
	Queues map[string]rmq.Queue
}

func NewRMQ(config Rcfg) *RMQConnection {
	poolSize := runtime.NumCPU() * 4
	errChan := make(chan error, 10)
	go logErrors(errChan)
	client := redis.NewClient(&redis.Options{
		Addr:            config.Address,
		Password:        config.Password,
		DB:              config.DB,
		PoolSize:        poolSize,
		PoolTimeout:     time.Duration(20) * time.Second,
		ReadTimeout:     time.Duration(20) * time.Second,
		WriteTimeout:    time.Duration(20) * time.Second,
		ConnMaxIdleTime: time.Duration(20) * time.Second,
	})
	connection, err := rmq.OpenConnectionWithRedisClient(tag, client, errChan)
	if err != nil {
		log.Fatal(err)
	}
	return &RMQConnection{
		RedisClient: client,
		Config:      config,
		Conn:        connection,
		Server: &RMQServer{
			Queues: make(map[string]rmq.Queue),
			conn:   connection,
		},
		Client: &RMQClient{
			Queues: make(map[string]rmq.Queue),
			conn:   connection,
		},
	}
}
func logErrors(errChan <-chan error) {
	for err := range errChan {
		switch err := err.(type) {
		case *rmq.HeartbeatError:
			if err.Count == rmq.HeartbeatErrorLimit {
				log.Print("heartbeat error (limit): ", err)
			} else {
				log.Print("heartbeat error: ", err)
			}
		case *rmq.ConsumeError:
			log.Print("consume error: ", err)
		case *rmq.DeliveryError:
			log.Print("delivery error: ", err.Delivery, err)
		default:
			log.Print("other error: ", err)
		}
	}
}

var (
	ERRQUEUEISEXIST    error = errors.New("queue is existed")
	ERRQUEUEISNOTEXIST error = errors.New("queue is not existed")
)

func (srv *RMQServer) AddQueue(name string, handler rmq.ConsumerFunc, numConsumers int) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if _, ok := srv.Queues[name]; ok {
		return ERRQUEUEISEXIST
	}
	queue, err := srv.conn.OpenQueue(name)
	if err != nil {
		return err
	}
	srv.Queues[name] = queue
	if err := queue.StartConsuming(prefetchLimit, pollDuration); err != nil {
		return err
	}
	for i := 0; i < numConsumers; i++ {
		if _, err := queue.AddConsumerFunc(tag, handler); err != nil {
			return err
		}
	}
	return nil
}

type Consumer struct {
}

func (c *Consumer) Consume(delivery rmq.Delivery) {
	log.Println("Received message: ", delivery.Payload())
	delivery.Ack()
}

func (conn *RMQConnection) Close() {
	<-conn.Conn.StopAllConsuming()
	cleaner := rmq.NewCleaner(conn.Conn)
	i, err := cleaner.Clean()
	if err != nil {
		log.Println(err)
	}
	log.Println("Cleaned", i, "messages")
}

func (srv *RMQServer) Run() {
	srv.waitForSignals()
}

func (srv *RMQServer) waitForSignals() {
	log.Println("Waiting for signals...")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()

	<-srv.conn.StopAllConsuming()
}

func (srv *RMQServer) RemoveQueue(name string) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if _, ok := srv.Queues[name]; !ok {
		return ERRQUEUEISNOTEXIST
	}
	q, err := srv.conn.OpenQueue(name)
	if err != nil {
		return err
	}
	q.Destroy()
	srv.Queues[name].StopConsuming()
	// srv.Queues[name].PurgeReady()
	// srv.Queues[name].PurgeRejected()
	delete(srv.Queues, name)
	return nil
}

func (srv *RMQServer) IsHasQueue(name string) bool {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	_, ok := srv.Queues[name]
	return ok
}

func (c *RMQClient) PublishBytes(queueName string, payload []byte) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	queue, ok := c.Queues[queueName]
	if !ok {
		queue, err = c.conn.OpenQueue(queueName)
		if err != nil {
			return err
		}
		c.Queues[queueName] = queue
		fmt.Println("new queue: ", queueName)
	}
	return queue.PublishBytes(payload)
}

func (c *RMQClient) Publish(queueName string, payload string) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	queue, ok := c.Queues[queueName]
	if !ok {
		queue, err = c.conn.OpenQueue(queueName)
		if err != nil {
			return err
		}
		c.Queues[queueName] = queue
		fmt.Println("new queue: ", queueName)
	}
	return queue.Publish(payload)
}
