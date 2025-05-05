package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	cache "github.com/besanh/chatting/common/caching"
	"github.com/besanh/chatting/common/constant"
	"github.com/besanh/chatting/common/util"
	"github.com/besanh/chatting/model"
	circuitbreaker "github.com/besanh/chatting/pkg/circuit_breaker"
	pkgOauth2 "github.com/besanh/chatting/pkg/oauth2"
	"github.com/besanh/chatting/repository"
	log "github.com/besanh/logger/logging/slog"
	"github.com/uptrace/bun"
	"golang.org/x/oauth2"
)

type (
	IUser interface {
		Login(ctx context.Context) (url string)
		OAuth2Callback(ctx context.Context, callbackData *model.OAuth2Callback) (token string, err error)
	}
	User struct {
		oAuth2Client pkgOauth2.IOAuth2
		userRepo     repository.IUser
	}
)

func NewUser(userRepo repository.IUser, oauthClient pkgOauth2.IOAuth2) IUser {
	return &User{
		userRepo:     userRepo,
		oAuth2Client: oauthClient,
	}
}

/*
 * Login and return url(google app)
 */
func (s *User) Login(ctx context.Context) (callbackUrl string) {
	// use PKCE to protect against CSRF attacks
	// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
	verifier := oauth2.GenerateVerifier()
	// Generate PKCE values.
	challenge := util.GenerateCodeChallenge(verifier)

	// Generate the authorization URL with PKCE parameters.
	authURL, err := url.Parse(s.oAuth2Client.AuthCodeUrl(OAUTH2_STATE, verifier))
	if err != nil {
		log.Error(err)
		return
	}

	q := authURL.Query()
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	authURL.RawQuery = q.Encode()

	// Caching
	redisKey := fmt.Sprintf("pkce:%s", OAUTH2_STATE)
	if err := cache.RCache.Set(redisKey, verifier, 3*time.Minute); err != nil {
		log.Error(err)
		return
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	callbackUrl = s.oAuth2Client.AuthCodeUrl(OAUTH2_STATE, verifier)
	return
}

func (s *User) OAuth2Callback(ctx context.Context, callbackData *model.OAuth2Callback) (token string, err error) {
	// PKCE
	verifierCache := cache.RCache.Get(fmt.Sprintf("pkce:%s", callbackData.State))
	if verifierCache == nil {
		err = fmt.Errorf("invalid state: %s", callbackData.State)
		return
	}
	verifier := verifierCache.(string)

	// Exchange the code for a token
	userInfo, err := s.oAuth2Client.Exchange(ctx, callbackData.Code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		log.Error(err)
		return
	}
	// Calculate TTL for the access token based on its expiry time
	ttl := time.Until(userInfo.Expiry)

	refreshTokenEncrypted, err := util.Encrypt(userInfo.RefreshToken)
	if err != nil {
		log.Error(err)
		return
	}

	user := model.User{
		GBase:  model.InitPgBase(),
		Status: constant.USER_STATUS_ACTIVE,
	}

	// Get user profile
	userProfile, err := s.getProfileUser(OAuth2Request{
		AccessToken: userInfo.AccessToken,
		Url:         GOOGLE_URL_USER_INFO,
		CBSetting: circuitbreaker.CBSetting{
			CBName:     "GOOGLE_USER_PROFILE",
			MaxRequest: 1,
			Interval:   5 * time.Second,
			TimeOut:    5 * time.Second,
			MaxTripCB:  3,
		},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		log.Error(err)
		return
	}
	user.UserProfile = userProfile

	// Marshal data and put to redis
	data, err := json.Marshal(user)
	if err != nil {
		log.Error(err)
		return
	}

	// Start transaction redis
	redisTx := cache.RCache.TxPineLine()
	cache.RCache.TxSet(ctx, redisTx, fmt.Sprintf("%s:%s", OAUTH2_TOKEN, userInfo.AccessToken), data, ttl)
	user.RefreshTokenEncrypted = refreshTokenEncrypted

	// Transaction
	if err = repository.DBConn.GetDB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) (err error) {
		_, total, err := s.userRepo.SelectByQuery(ctx, repository.DBConn, []model.Param{
			{
				Key:      "user_profile->>'email'",
				Operator: "=",
				Value:    userProfile.Email,
			},
		}, 1, 0)
		if err != nil {
			return
		} else if total == 0 {
			if err = s.userRepo.TxInsert(ctx, tx, user); err != nil {
				return
			}
		}

		return
	}); err != nil {
		log.Error(err)
		return
	}

	// Commit redis
	if _, err = redisTx.Exec(ctx); err != nil {
		log.Error(err)
		return
	}

	token = userInfo.AccessToken

	return
}
