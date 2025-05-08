package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/besanh/chatting/common/caching"
	"github.com/besanh/chatting/config"
	"github.com/besanh/chatting/model"
	"github.com/besanh/chatting/service"
	log "github.com/besanh/logger/logging/slog"
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type WsHandler struct {
	ctx        context.Context
	cancel     context.CancelFunc
	cfg        config.Config
	subscriber service.ISubscriber
}

type WsSession struct {
	Parent          *WsHandler          `json:"-"`
	CtxWs           context.Context     `json:"-"`
	Cancel          context.CancelFunc  `json:"-"`
	Con             *websocket.Conn     `json:"-"`
	MsgReadChannel  chan []byte         `json:"-"`
	MsgWriteChannel chan []byte         `json:"-"`
	ErrChannel      chan error          `json:"-"`
	WsEvent         chan *model.WsEvent `json:"-"`
	UserId          string              `json:"user_id"`
	SessionId       string              `json:"session_id"`
	CountMessage    uint64              `json:"count_message"`
}

func NewWs(r *gin.Engine, cfg config.Config, subscriberService service.ISubscriber) *WsHandler {
	ctx, cancel := context.WithCancel(context.Background())
	handler := &WsHandler{
		ctx:        ctx,
		cancel:     cancel,
		cfg:        cfg,
		subscriber: subscriberService,
	}

	service.WsSubscribers = &service.Subscribers{
		SubscriberMessageBuffer: 16,
		Subscribers:             make(map[*service.Subscriber]struct{}),
		PublishLimiter:          rate.NewLimiter(rate.Every(time.Millisecond*100), 100),
	}

	group := r.Group("chatting/ws/v1")
	{
		group.GET("subscribe/:user_id/:session_id", handler.Subscribe)
	}

	return handler
}

func (handler *WsHandler) Subscribe(c *gin.Context) {
	userID := c.Param("user_id")
	sessionID := c.Param("session_id")
	if userID == "" || sessionID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing user_id or session_id"})
		return
	}

	// Rate limiting
	rlKey := fmt.Sprintf("rate:ws:sub:%s", userID)
	cnt, err := caching.RCache.Incr(c, rlKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate-limit error"})
		return
	}
	if cnt == 1 {
		caching.RCache.Expire(c, rlKey, time.Second)
	}
	if cnt > int64(handler.cfg.Api.RateLimit) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithCancel(handler.ctx)

	// Create and register session
	sessCtx := &WsSession{
		Parent:          handler,
		CtxWs:           ctx,
		Cancel:          cancel,
		Con:             conn,
		MsgReadChannel:  make(chan []byte),
		MsgWriteChannel: make(chan []byte),
		ErrChannel:      make(chan error),
		WsEvent:         make(chan *model.WsEvent, 1),
		UserId:          userID,
		SessionId:       sessionID,
	}

	// Store session in Redis
	if err := storeSessionRedis(c, userID, sessionID); err != nil {
		log.Warn("failed to store session in Redis: ", err)
	}

	// Subscribe to JetStream
	if err := handler.subscriber.Subscribe(ctx, userID, sessionID, sessCtx.WsEvent); err != nil {
		cancel()
		return
	}

	go sessCtx.readPump()
	go sessCtx.writePump()

	<-ctx.Done()

	// Cleanup Redis
	deleteSessionRedis(userID, sessionID)
}

func storeSessionRedis(ctx *gin.Context, userID, sessionID string) error {
	field := fmt.Sprintf("%s:%s", userID, sessionID)
	data := map[string]any{
		"user_id":      userID,
		"session_id":   sessionID,
		"connected_at": time.Now().Format(time.RFC3339),
		"user_agent":   ctx.Request.UserAgent(),
		"remote_addr":  ctx.Request.RemoteAddr,
	}

	byteData, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return err
	}

	values := []any{field, string(byteData)}
	err = caching.RCache.HSet(service.CHATTING_CONNECTION_KEY, values)
	if err != nil {
		log.Error("failed to store session in Redis hash: ", err)
	}
	return err
}

func deleteSessionRedis(userID, sessionID string) {
	key := fmt.Sprintf("ws:user_sessions:%s", userID)
	if err := caching.RCache.HDel(key, sessionID); err != nil {
		log.Warn("failed to delete session from Redis hash: ", err)
	}
}

// readPump just listens for client messages (you can apply per-message rate-limit here)
func (s *WsSession) readPump() {
	defer s.Cancel()

	for {
		// Use your session context to bound the Read
		_, msg, err := s.Con.Read(s.CtxWs)
		if err != nil {
			// send the error into your error channel
			s.ErrChannel <- err
			return
		}
		log.Info(msg)
	}
}

// writePump pushes server‐side events down to the client
func (s *WsSession) writePump() {
	defer s.Cancel()
	for {
		select {
		case evt := <-s.WsEvent:
			b, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			// Send a text message using the session’s context
			if err := s.Con.Write(s.CtxWs, websocket.MessageText, b); err != nil {
				s.ErrChannel <- err
				return
			}
		case err := <-s.ErrChannel:
			log.Error(err)
			return
		case <-s.CtxWs.Done():
			return
		}
	}
}
