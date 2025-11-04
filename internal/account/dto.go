package account

// RegisterRequest - payload for user registration
type RegisterRequest struct {
	Email     string `json:"email" example:"user@example.com"`
	Password  string `json:"password" example:"strongpassword"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
}

// LoginRequest - payload for user login
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"strongpassword"`
}

// ProfileResponse - payload returned by /profile
type ProfileResponse struct {
	ID        int    `json:"id" example:"1"`
	Email     string `json:"email" example:"user@example.com"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
	Role      string `json:"role" example:"student"`
}

// ChangePasswordRequest - payload for changing password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" example:"currentpass"`
	NewPassword string `json:"new_password" example:"newpass123"`
}

// UpdateProfileRequest - payload for updating profile
type UpdateProfileRequest struct {
	Email     string `json:"email" example:"newemail@example.com"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
}

// SetRoleRequest - payload for admin to set user role
type SetRoleRequest struct {
	UserID int    `json:"user_id" example:"7"`
	Role   string `json:"role" example:"teacher"`
}
