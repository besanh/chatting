package v1

import (
	"github.com/besanh/chatting/common/response"
	"github.com/besanh/chatting/middleware/oauth2"
	"github.com/besanh/chatting/model"
	"github.com/besanh/chatting/service"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	ChatService service.IChat
}

func NewChat(engine *gin.Engine, chatService service.IChat) {
	handler := &ChatHandler{
		ChatService: chatService,
	}

	group := engine.Group("chatting/chat/v1").Use(oauth2.NewOAuth2Middleware())
	{
		// group.GET("get/:chat_id", handler.GetChatById)
		// group.GET("list", handler.GetChats)
		group.POST("create", handler.InsertChat)
		// group.POST("update/:chat_id", handler.UpdateChatById)
		// group.POST("delete/:chat_id", handler.DeleteChatById)
	}
}

func (handler *ChatHandler) InsertChat(c *gin.Context) {
	request := &model.PinelineCreateChatRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	if err := request.Validate(); err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	id, err := handler.ChatService.InsertChat(c, *request)
	if err != nil {
		c.JSON(response.ServiceUnavailableMsg(err.Error()))
		return
	}

	c.JSON(response.OK(map[string]string{
		"id": id,
	}))
}
