package domain

import (
	"context"
	"time"
)

const AdminDomain = "admin"

type Admin struct {
	UserNationalID NationalID `json:"user_national_id"`
	Email          Email      `json:"email"`
	CreatedAt      time.Time  `json:"created_at"`
}

type AdminStorage interface {
	Store(context.Context, Admin) error
	Find(context.Context, NationalID) (Admin, error)
	Delete(context.Context, NationalID) error
}

type AdminService interface {
	Create(context.Context, Admin) error
	Get(context.Context, NationalID) (Admin, error)
	Delete(context.Context, NationalID) error
}
