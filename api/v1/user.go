package v1

import (
	"github.com/besanh/chatting/common/response"
	"github.com/besanh/chatting/middleware/oauth2"
	"github.com/besanh/chatting/model"
	"github.com/besanh/chatting/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.IUser
}

func NewUsers(engine *gin.Engine, userService service.IUser) {
	handler := &UserHandler{
		userService: userService,
	}

	group := engine.Group("v1/user")
	{
		group.GET("login", handler.Login)
		group.GET("oauth2callback", handler.OAuth2Callback)
		group.Use(oauth2.NewOAuth2Middleware()).GET("me", handler.Me)
	}
}

func (handler *UserHandler) Login(c *gin.Context) {
	url := handler.userService.Login(c)
	if len(url) < 1 {
		c.JSON(response.BadRequestMsg("url is empty"))
		return
	}

	c.JSON(response.OK(map[string]any{
		"url": url,
	}))
}

func (handler *UserHandler) OAuth2Callback(c *gin.Context) {
	code := c.Query("code")
	if len(code) < 1 {
		c.JSON(response.BadRequestMsg("code is empty"))
		return
	}

	state := c.Query("state")
	if len(state) < 1 {
		c.JSON(response.BadRequestMsg("state is empty"))
		return
	}

	scope := c.Query("scope")
	if len(scope) < 1 {
		c.JSON(response.BadRequestMsg("scope is empty"))
		return
	}

	callbackData := &model.OAuth2Callback{
		Code:  code,
		State: state,
		Scope: scope,
	}

	token, err := handler.userService.OAuth2Callback(c, callbackData)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	c.JSON(response.Created(map[string]any{
		"token": token,
	}))
}

func (handler *UserHandler) Me(c *gin.Context) {
	user, err := oauth2.GetUser(c)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	c.JSON(response.OK(user))
}
