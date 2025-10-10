package domain

import "context"

const AuthhenticationDomain = "authentication"

type AuthToken string

type LoginRequest struct {
	Usrname  Username `json:"username"`
	Password Password `json:"password"`
}

type LoginResponse struct {
	Token AuthToken `json:"token"`
}

type RegisterRequest struct {
	NationalID   NationalID   `json:"national_id"`
	FirstName    string       `json:"first_name"`
	LastName     string       `json:"last_name"`
	Username     Username     `json:"username"`
	Password     Password     `json:"password"`
	Email        Email        `json:"email"`
	MobileNumber MobileNumber `json:"mobile_number"`
	Gender       string       `json:"gender"`
	Role         string       `json:"role"`
}

type RegisterResponse struct {
	User  User      `json:"user"`
	Token AuthToken `json:"token"`
}

type AuthenticationService interface {
	Login(context.Context, LoginRequest) (LoginResponse, error)
	Register(context.Context, RegisterRequest) (RegisterResponse, error)
	Logout(context.Context) error
	SendResetPasswordToken(context.Context, Username, NationalID) error
	ResetPassword(ctx context.Context, username Username, nationalId NationalID, token string, newPassword string) error
}
