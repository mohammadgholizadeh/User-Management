package service

import (
	"context"

	"github.com/mohammadgholizadeh/user-management/internal/domain"
)

type admin struct {
	adminStorage domain.AdminStorage
}

func NewAdminService(adminStorage domain.AdminStorage) domain.AdminService {
	return &admin{adminStorage: adminStorage}
}

func (a *admin) Create(ctx context.Context, admin domain.Admin) error {
	return a.adminStorage.Store(ctx, admin)
}

func (a *admin) Get(ctx context.Context, id domain.NationalID) (domain.Admin, error) {
	return a.adminStorage.Find(ctx, id)
}

func (a *admin) Delete(ctx context.Context, id domain.NationalID) error {
	return a.adminStorage.Delete(ctx, id)
}
