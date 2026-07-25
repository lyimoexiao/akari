package auth

// RegisterReq contains the fields needed to register a new user.
type RegisterReq struct {
	Username   string         `json:"username" binding:"required,min=3,max=64"`
	Email      string         `json:"email" binding:"required,email"`
	Password   string         `json:"password" binding:"required,min=6,max=128"`
	CaptchaID  string         `json:"captcha_id,omitempty"`
	UserAnswer map[string]any `json:"user_answer,omitempty"`
}

// LoginReq contains login credentials.
type LoginReq struct {
	Username   string         `json:"username" binding:"required"`
	Password   string         `json:"password" binding:"required"`
	CaptchaID  string         `json:"captcha_id,omitempty"`
	UserAnswer map[string]any `json:"user_answer,omitempty"`
}

// VerifyEmailReq contains the token for email verification.
type VerifyEmailReq struct {
	Token string `json:"token" binding:"required"`
}

// ForgotPasswordReq contains the email for requesting a password reset.
type ForgotPasswordReq struct {
	Email      string         `json:"email" binding:"required,email"`
	CaptchaID  string         `json:"captcha_id,omitempty"`
	UserAnswer map[string]any `json:"user_answer,omitempty"`
}

// ResetPasswordReq contains the token and new password for resetting a password.
type ResetPasswordReq struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}
