package account

// RegisterRequest - payload for user registration
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// LoginRequest - payload for user login
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"strongpassword"`
}

// LoginResponse - payload returned by /login

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ProfileResponse - payload returned by /profile
type ProfileResponse struct {
	ID        int    `json:"id" example:"1"`
	UID       string `json:"uid" example:"123456"`
	Email     string `json:"email" example:"user@example.com"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
	GroupName string `json:"group_name" example:"Group 1234"`
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

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
