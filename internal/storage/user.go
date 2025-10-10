package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/user-management/internal/domain"
)

func genderToSmallInt(g string) int16 {
	switch strings.ToLower(strings.TrimSpace(g)) {
	case "male":
		return 1
	case "female":
		return 2
	default:
		return 0
	}
}

func smallIntToGender(v int16) string {
	switch v {
	case 1:
		return "male"
	case 2:
		return "female"
	default:
		return "unknown"
	}
}

type user struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewUserStore(pool *pgxpool.Pool, logger *zap.Logger) domain.UserStore {
	return &user{pool: pool, logger: logger}
}

func (s *user) Store(ctx context.Context, u domain.User) error {
	if u.NationalID == "" {
		return fmt.Errorf("national_id is required")
	}
	if u.UserName == "" {
		return fmt.Errorf("username is required")
	}
	if u.HashedPassword == "" {
		return fmt.Errorf("hashed_password is required")
	}
	if u.Role == "" {
		u.Role = "user"
	}

	query := `
		INSERT INTO users (
			national_id, first_name, last_name, username, mobile_number, gender, email, role, hashed_password, is_active
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`
	_, err := s.pool.Exec(ctx, query,
		u.NationalID,
		u.FirstName,
		u.LastName,
		u.UserName,
		u.MobileNumber,
		genderToSmallInt(u.Gender),
		u.Email,
		u.Role,
		u.HashedPassword,
		u.IsActive,
	)
	if err != nil {
		s.logger.Error("user.store failed",
			zap.String("domain", "user"),
			zap.String("national_id", string(u.NationalID)),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (s *user) FindAllUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	query := `
		SELECT 
			national_id, first_name, last_name, username, mobile_number, gender, email, role, hashed_password, is_active,
			created_at, COALESCE(updated_at, created_at)
		FROM users
		ORDER BY created_at DESC
		OFFSET $1
		LIMIT $2
	`
	rows, err := s.pool.Query(ctx, query, offset, limit)
	if err != nil {
		s.logger.Error("user.find_all query failed", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var g int16
		if err := rows.Scan(
			&u.NationalID, &u.FirstName, &u.LastName, &u.UserName, &u.MobileNumber, &g, &u.Email, &u.Role, &u.HashedPassword, &u.IsActive,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			s.logger.Error("user.find_all scan failed", zap.Error(err))
			return nil, err
		}
		u.Gender = smallIntToGender(g)
		users = append(users, u)
	}
	return users, nil
}

func (s *user) FindUserByNationalID(ctx context.Context, id domain.NationalID) (domain.User, error) {
	query := `
		SELECT 
			national_id, first_name, last_name, username, mobile_number, gender, email, role, hashed_password, is_active,
			created_at, COALESCE(updated_at, created_at)
		FROM users WHERE national_id = $1
	`
	row := s.pool.QueryRow(ctx, query, id)
	var u domain.User
	var g int16
	if err := row.Scan(
		&u.NationalID, &u.FirstName, &u.LastName, &u.UserName, &u.MobileNumber, &g, &u.Email, &u.Role, &u.HashedPassword, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		s.logger.Warn("user not found by national_id", zap.String("national_id", string(id)), zap.Error(err))
		return domain.User{}, err
	}
	u.Gender = smallIntToGender(g)
	return u, nil
}

func (s *user) FindUserByUsername(ctx context.Context, username domain.Username) (domain.User, error) {
	query := `
		SELECT 
			national_id, first_name, last_name, username, mobile_number, gender, email, role, hashed_password, is_active,
			created_at, COALESCE(updated_at, created_at)
		FROM users WHERE username = $1
	`
	row := s.pool.QueryRow(ctx, query, username)
	var u domain.User
	var g int16
	if err := row.Scan(
		&u.NationalID, &u.FirstName, &u.LastName, &u.UserName, &u.MobileNumber, &g, &u.Email, &u.Role, &u.HashedPassword, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		s.logger.Warn("user not found by username", zap.String("username", string(username)), zap.Error(err))
		return domain.User{}, err
	}
	u.Gender = smallIntToGender(g)
	return u, nil
}

func (s *user) FindUsersByRole(ctx context.Context, role string) ([]domain.User, error) {
	query := `
		SELECT 
			national_id, first_name, last_name, username, mobile_number, gender, email, role, hashed_password, is_active,
			created_at, COALESCE(updated_at, created_at)
		FROM users WHERE role = $1
		ORDER BY created_at DESC
	`
	rows, err := s.pool.Query(ctx, query, role)
	if err != nil {
		s.logger.Error("user.find_by_role query failed", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var g int16
		if err := rows.Scan(
			&u.NationalID, &u.FirstName, &u.LastName, &u.UserName, &u.MobileNumber, &g, &u.Email, &u.Role, &u.HashedPassword, &u.IsActive,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			s.logger.Error("user.find_by_role scan failed", zap.Error(err))
			return nil, err
		}
		u.Gender = smallIntToGender(g)
		users = append(users, u)
	}
	return users, nil
}

func (s *user) FindUserRole(ctx context.Context, id domain.NationalID) (domain.User, error) {
	row := s.pool.QueryRow(ctx, `SELECT national_id, role FROM users WHERE national_id = $1`, id)
	var u domain.User
	if err := row.Scan(&u.NationalID, &u.Role); err != nil {
		s.logger.Warn("user role not found", zap.String("national_id", string(id)), zap.Error(err))
		return domain.User{}, err
	}
	return u, nil
}

func (s *user) FindUserByMobileNumber(ctx context.Context, m domain.MobileNumber) (domain.User, error) {
	query := `
		SELECT 
			national_id, first_name, last_name, username, mobile_number, gender, email, role, hashed_password, is_active,
			created_at, COALESCE(updated_at, created_at)
		FROM users WHERE mobile_number = $1
	`
	row := s.pool.QueryRow(ctx, query, m)
	var u domain.User
	var g int16
	if err := row.Scan(
		&u.NationalID, &u.FirstName, &u.LastName, &u.UserName, &u.MobileNumber, &g, &u.Email, &u.Role, &u.HashedPassword, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		s.logger.Warn("user not found by mobile_number", zap.String("mobile_number", string(m)), zap.Error(err))
		return domain.User{}, err
	}
	u.Gender = smallIntToGender(g)
	return u, nil
}

func (s *user) Update(ctx context.Context, u domain.User, id domain.NationalID) error {
	if u.HashedPassword == "" {
		row := s.pool.QueryRow(ctx, `SELECT hashed_password FROM users WHERE national_id = $1`, id)
		var cur string
		if err := row.Scan(&cur); err != nil {
			s.logger.Error("user.update: fetch current hash failed", zap.Error(err))
			return err
		}
		u.HashedPassword = cur
	}

	query := `
		UPDATE users SET
			first_name = $1,
			last_name = $2,
			username = $3,
			mobile_number = $4,
			gender = $5,
			email = $6,
			role = $7,
			hashed_password = $8,
			is_active = $9,
			updated_at = $10
		WHERE national_id = $11
	`
	_, err := s.pool.Exec(ctx, query,
		u.FirstName,
		u.LastName,
		u.UserName,
		u.MobileNumber,
		genderToSmallInt(u.Gender),
		u.Email,
		u.Role,
		u.HashedPassword,
		u.IsActive,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		s.logger.Error("user.update failed",
			zap.String("national_id", string(id)),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (s *user) Delete(ctx context.Context, id domain.NationalID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM users WHERE national_id = $1`, id)
	if err != nil {
		s.logger.Error("user.delete failed", zap.String("national_id", string(id)), zap.Error(err))
		return err
	}
	return nil
}
