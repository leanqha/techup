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
