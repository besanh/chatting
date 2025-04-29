package model

type WsEvent struct {
	EventType string `json:"event_type"`
	Data      struct {
		Message string `json:"message"`
		MsgId   string `json:"msg_id"`
	} `json:"data"`
}
