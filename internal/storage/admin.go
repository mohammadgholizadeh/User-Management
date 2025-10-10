package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/user-management/internal/domain"
)

type admin struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewAdminStorage(pool *pgxpool.Pool, logger *zap.Logger) domain.AdminStorage {
	return &admin{pool: pool, logger: logger}
}

func (s *admin) Store(ctx context.Context, a domain.Admin) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO admins (user_national_id, email, created_at) VALUES ($1,$2,NOW())`,
		a.UserNationalID, a.Email,
	)
	if err != nil {
		s.logger.Error("admin.store failed", zap.String("national_id", string(a.UserNationalID)), zap.Error(err))
		return err
	}
	return nil
}

func (s *admin) Find(ctx context.Context, id domain.NationalID) (domain.Admin, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT user_national_id, email, created_at FROM admins WHERE user_national_id = $1`, id,
	)
	var a domain.Admin
	if err := row.Scan(&a.UserNationalID, &a.Email, &a.CreatedAt); err != nil {
		s.logger.Warn("admin.find not found", zap.String("national_id", string(id)), zap.Error(err))
		return domain.Admin{}, err
	}
	return a, nil
}

func (s *admin) Delete(ctx context.Context, id domain.NationalID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM admins WHERE user_national_id = $1`, id)
	if err != nil {
		s.logger.Error("admin.delete failed", zap.String("national_id", string(id)), zap.Error(err))
		return err
	}
	return nil
}
