package admin

// ListUsersReq holds query parameters for listing users.
type ListUsersReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Query    string `form:"query,omitempty"` // search by username or email
}

// AdminVerifyEmailReq is sent to manually verify a user's email.
type AdminVerifyEmailReq struct {
	UserID uint `json:"user_id" binding:"required"`
}

// SetRoleReq is sent to change a user's role.
type SetRoleReq struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=super_admin staff user"`
}

// AdminResetPasswordReq is sent to forcibly reset a user's password.
type AdminResetPasswordReq struct {
	UserID      uint   `json:"user_id" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=128"`
}

// DeleteUserReq identifies the user to delete.
type DeleteUserReq struct {
	UserID uint `json:"user_id" binding:"required"`
}
