package oauth2

import (
	"fmt"
	"net/http"
	"strings"

	cache "github.com/besanh/chatting/common/caching"
	"github.com/besanh/chatting/common/util"
	"github.com/besanh/chatting/model"
	"github.com/besanh/chatting/service"
	log "github.com/besanh/logger/logging/slog"
	"github.com/gin-gonic/gin"
)

func NewOAuth2Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		headerValue := c.GetHeader("Authorization")
		token := parseTokenFromAuthorization(headerValue)
		user, err := validateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{
				"error": err.Error(),
			})
			return
		}
		c.Set("USER", user)
		c.Next()
	}
}

func parseTokenFromAuthorization(authorizationHeader string) string {
	return strings.Replace(authorizationHeader, "Bearer ", "", 1)
}

func validateToken(tokenString string) (userInfo *model.User, err error) {
	// Because the token is stored in redis, so I need to get the user info from it
	dataCache := cache.RCache.Get(fmt.Sprintf("%s:%s", service.OAUTH2_TOKEN, tokenString))
	if dataCache == nil {
		err = fmt.Errorf("invalid token")
		log.Error(err)
		return
	}

	if err = util.ParseAnyToAny(dataCache, &userInfo); err != nil {
		log.Error(err)
		return
	}

	return
}

func GetUser(c *gin.Context) (result *model.UserResponse, err error) {
	user, exist := c.Get("USER")
	if !exist {
		err = fmt.Errorf("user not found")
		log.Error(err)
		return
	}

	if err = util.ParseAnyToAny(user, &result); err != nil {
		log.Error(err)
		return
	}

	return
}
