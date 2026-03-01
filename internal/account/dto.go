package account

// RegisterRequest describes payload for account registration.
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// LoginRequest describes payload for account authentication.
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"strongpassword"`
}

// LoginResponse represents token payloads (when returned in JSON).
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ProfileResponse represents the authenticated user's profile.
type ProfileResponse struct {
	ID         int    `json:"id" example:"1"`
	UID        string `json:"uid" example:"123456"`
	Email      string `json:"email" example:"user@example.com"`
	FirstName  string `json:"first_name" example:"John"`
	MiddleName string `json:"middle_name" example:"Jane"`
	LastName   string `json:"last_name" example:"Doe"`
	GroupID    int    `json:"group_id" example:"1"`
	GroupName  string `json:"group_name" example:"Group 1234"`
	Role       string `json:"role" example:"student"`
}

// ChangePasswordRequest describes payload for password change.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" example:"currentpass"`
	NewPassword string `json:"new_password" example:"newpass123"`
}

// UpdateProfileRequest describes payload for profile updates.
type UpdateProfileRequest struct {
	Email     string `json:"email" example:"newemail@example.com"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
}

// SetRoleRequest describes payload for assigning a role to a user.
type SetRoleRequest struct {
	UserID int    `json:"user_id" example:"7"`
	Role   string `json:"role" example:"teacher"`
}

// RefreshRequest describes payload for manual refresh token submission.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AdminAccountsFilter describes filters for admin account listing.
type AdminAccountsFilter struct {
	Role       *string
	GroupID    *int
	Email      *string
	Name       *string
	UID        *string
	IsVerified *bool
}

// AdminUpdateAccountRequest describes payload for admin account updates.
type AdminUpdateAccountRequest struct {
	Email      *string `json:"email"`
	FirstName  *string `json:"first_name"`
	MiddleName *string `json:"middle_name"`
	LastName   *string `json:"last_name"`
	Role       *string `json:"role"`
	GroupID    *int    `json:"group_id"`
	IsVerified *bool   `json:"is_verified"`
}

// AdminAccountResponse represents an account payload for admin tools.
type AdminAccountResponse struct {
	ID         int    `json:"id"`
	UID        string `json:"uid"`
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	Role       string `json:"role"`
	IsVerified bool   `json:"is_verified"`
	GroupID    int    `json:"group_id"`
	GroupName  string `json:"group_name"`
}

// ForgotPasswordRequest describes payload for initiating a password reset.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest describes payload for completing a password reset.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
