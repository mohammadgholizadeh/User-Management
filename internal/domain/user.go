package domain

import (
	"context"
	"time"
)

type User struct {
	NationalID     NationalID   `json:"national_id"`
	FirstName      string       `json:"first_name"`
	LastName       string       `json:"last_name"`
	Password       Password     `json:"-"`
	HashedPassword string       `json:"-"`
	UserName       Username     `json:"username"`
	Role           string       `json:"role"`
	MobileNumber   MobileNumber `json:"mobile_number"`
	Gender         string       `json:"gender"`
	Email          Email        `json:"email"`
	IsActive       bool         `json:"is_active"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type UserStore interface {
	Store(context.Context, User) error
	FindAllUsers(ctx context.Context, limit, offset int) ([]User, error)
	FindUserByNationalID(context.Context, NationalID) (User, error)
	FindUserByUsername(context.Context, Username) (User, error)
	FindUsersByRole(ctx context.Context, Role string) ([]User, error)
	FindUserRole(context.Context, NationalID) (User, error)
	FindUserByMobileNumber(context.Context, MobileNumber) (User, error)
	Update(context.Context, User, NationalID) error
	Delete(context.Context, NationalID) error
}

type UserService interface {
	Create(context.Context, User) error
	GetAllUsers(ctx context.Context, limit, offset int) ([]User, error)
	GetUserByNationalID(context.Context, NationalID) (User, error)
	GetUserByUsername(context.Context, Username) (User, error)
	GetUsersByRole(ctx context.Context, Role string) ([]User, error)
	GetUserRole(context.Context, NationalID) (User, error)
	GetUserByMobileNumber(context.Context, MobileNumber) (User, error)
	Update(context.Context, User, NationalID) error
	Delete(context.Context, NationalID) error
}
