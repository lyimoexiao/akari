package user

type ListUsersReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Query    string `form:"query,omitempty"`
}

type VerifyEmailReq struct {
	UserID uint `json:"user_id" binding:"required"`
}

type ResetPasswordReq struct {
	UserID      uint   `json:"user_id" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=128"`
}

type DeleteUserCommand struct {
	CallerID uint
	UserID   uint
}
