package model

type (
	User struct {
		*GBase
		UserProfile           UserProfile `json:"user_profile" bson:"user_profile"`
		RefreshTokenEncrypted string      `json:"refresh_token_encrypted" bson:"refresh_token_encrypted"`
		Status                string      `json:"status" bson:"status"`
		Scope                 []string    `json:"scope" bson:"scope"`
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
		UserProfile UserProfile `json:"user_profile"`
		Status      string      `json:"status"`
		Scope       []string    `json:"scope"`
	}
)
