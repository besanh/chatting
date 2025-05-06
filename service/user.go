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
	"github.com/besanh/chatting/config"
	"github.com/besanh/chatting/model"
	circuitbreaker "github.com/besanh/chatting/pkg/circuit_breaker"
	pkgOauth2 "github.com/besanh/chatting/pkg/oauth2"
	"github.com/besanh/chatting/repository"
	log "github.com/besanh/logger/logging/slog"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"golang.org/x/oauth2"
)

type (
	IUser interface {
		Login(ctx context.Context) (url string)
		OAuth2Callback(ctx context.Context, callbackData *model.OAuth2Callback) (token string, err error)
		Logout(ctx *gin.Context, refreshTokenEncrypted, id string) (err error)
	}
	User struct {
		oAuth2Client pkgOauth2.IOAuth2
		userRepo     repository.IUser
		config       config.Config
	}
)

func NewUser(userRepo repository.IUser, oauthClient pkgOauth2.IOAuth2, cfg config.Config) IUser {
	return &User{
		userRepo:     userRepo,
		oAuth2Client: oauthClient,
		config:       cfg,
	}
}

/*
 * Login and return url(google app)
 */
func (s *User) Login(ctx context.Context) (callbackUrl string) {
	// 1. Generate PKCE verifier + challenge
	verifier := oauth2.GenerateVerifier()
	challenge := util.GenerateCodeChallenge(verifier)

	// 2. Build the base auth URL (only state & verifier)
	rawURL := s.oAuth2Client.AuthCodeUrl(OAUTH2_STATE, verifier)
	authURL, err := url.Parse(rawURL)
	if err != nil {
		log.Error(err)
		return
	}

	// 3. Inject all PKCE + offline/consent parameters
	q := authURL.Query()
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("access_type", "offline") // request a refresh token
	q.Set("prompt", "consent")      // force re-consent so Google returns it
	authURL.RawQuery = q.Encode()

	// 4. Cache the verifier for your PKCE check
	redisKey := fmt.Sprintf("pkce:%s", OAUTH2_STATE)
	if err := cache.RCache.Set(redisKey, verifier, 3*time.Minute); err != nil {
		log.Error(err)
		return
	}

	// 5. Return the full URL
	callbackUrl = authURL.String()
	return
}

func (s *User) OAuth2Callback(ctx context.Context, callbackData *model.OAuth2Callback) (token string, err error) {
	// 1. Verify PKCE state
	raw := cache.RCache.Get(fmt.Sprintf("pkce:%s", callbackData.State))
	if raw == nil {
		err = fmt.Errorf("invalid state: %s", callbackData.State)
		log.Error(err)
		return
	}
	verifier := raw.(string)

	// 2. Exchange code for tokens
	userInfo, err := s.oAuth2Client.Exchange(ctx, callbackData.Code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		log.Error(err)
		return
	}
	ttl := time.Until(userInfo.Expiry)

	// 3. Encrypt only if non-empty (Google only sends refresh_token on first consent)
	var newEnc string
	if userInfo.RefreshToken != "" {
		newEnc, err = util.Encrypt(userInfo.RefreshToken)
		if err != nil {
			log.Error(err)
			return
		}
	}

	// 4. Build your model.User
	user := model.User{
		GBase:  model.InitPgBase(),
		Status: constant.USER_STATUS_ACTIVE,
	}
	profile, err := s.getProfileUser(OAuth2Request{
		AccessToken: userInfo.AccessToken,
		Url:         s.config.Pkg.Oauth2.Google.UserInfoUrl,
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
	user.UserProfile = profile

	// 5. Upsert into Postgres, preserving existing encrypted token if none returned
	if err = repository.DBConn.GetDB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		users, total, err := s.userRepo.SelectByQuery(ctx, repository.DBConn,
			[]model.Param{{
				Key:      "user_profile->>'email'",
				Operator: "=",
				Value:    profile.Email,
			},
				{
					Key:      "status",
					Operator: "=",
					Value:    constant.USER_STATUS_ACTIVE,
				},
			},
			1, 0,
		)
		if err != nil {
			return err
		}

		if total == 0 {
			// first-time: must have gotten a refresh_token
			user.RefreshTokenEncrypted = newEnc
			return s.userRepo.TxInsert(ctx, tx, user)
		}

		// existing user
		existing := (*users)[0]
		user.Id, user.GBase = existing.Id, existing.GBase

		if newEnc != "" {
			user.RefreshTokenEncrypted = newEnc
		} else {
			user.RefreshTokenEncrypted = existing.RefreshTokenEncrypted
		}
		return s.userRepo.TxUpdate(ctx, tx, user)
	}); err != nil {
		log.Error(err)
		return
	}

	// 6. Store into Redis (now with the correct encrypted token)
	buf, err := json.Marshal(user)
	if err != nil {
		log.Error(err)
		return
	}
	redisTx := cache.RCache.TxPineLine()
	cache.RCache.TxSet(ctx, redisTx,
		fmt.Sprintf("%s:%s", OAUTH2_TOKEN, userInfo.AccessToken),
		buf, ttl,
	)
	if _, err = redisTx.Exec(ctx); err != nil {
		log.Error(err)
		return
	}

	// 7. Return the new access token
	token = userInfo.AccessToken
	return
}

func (s *User) Logout(ctx *gin.Context, refreshTokenEncrypted, id string) (err error) {
	err = s.revokeGoogleToken(OAuth2Request{
		RefreshToken: refreshTokenEncrypted,
		Url:          s.config.Pkg.Oauth2.Google.RevokeUrl,
		CBSetting: circuitbreaker.CBSetting{
			CBName:     "GOOGLE_REVOKE_TOKEN",
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

	if err = repository.UserRepo.ClearRefreshToken(ctx, id); err != nil {
		log.Error(err)
		return
	}

	return
}
