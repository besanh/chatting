package model

type (
	User struct {
		*GBase
		UserProfile           UserProfile `json:"user_profile" bun:"user_profile,type:jsonb,notnull"`
		RefreshTokenEncrypted string      `json:"refresh_token_encrypted,omitempty" bun:"refresh_token_encrypted,type:text"`
		Status                string      `json:"status" bun:"status,type:varchar(50),notnull"`
		Scope                 []string    `json:"scope" bun:"scope,type:text[]"`
	}

	UserProfile struct {
		Sub           string `json:"sub"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
		Profile       string `json:"profile"`
		Email         string `json:"email,omitempty"`
		EmailVerified bool   `json:"email_verified,omitempty"`
	}

	UserResponse struct {
		*GBase
		UserProfile           UserProfile `json:"user_profile"`
		RefreshTokenEncrypted string      `json:"refresh_token_encrypted,omitempty"`
		Status                string      `json:"status"`
		Scope                 []string    `json:"scope"`
	}
)
