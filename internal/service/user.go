package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/user-management/internal/domain"
)

type user struct {
	store  domain.UserStore
	logger *zap.Logger
}

func NewUserService(store domain.UserStore, logger *zap.Logger) domain.UserService {
	return &user{store: store, logger: logger}
}

func (s *user) Create(ctx context.Context, u domain.User) error {
	if err := u.NationalID.Validate(); err != nil {
		s.logger.Warn("invalid national id", zap.Error(err))
		return err
	}
	if err := u.UserName.Validate(); err != nil {
		s.logger.Warn("invalid username", zap.Error(err))
		return err
	}
	if err := u.Email.Validate(); err != nil {
		s.logger.Warn("invalid email", zap.Error(err))
		return err
	}
	if err := u.Password.Validate(); err != nil {
		s.logger.Warn("invalid password", zap.Error(err))
		return err
	}
	if err := u.MobileNumber.Validate(); err != nil {
		s.logger.Warn("invalid mobile number", zap.Error(err))
		return err
	}

	hash, err := u.Password.Hash()
	if err != nil {
		s.logger.Error("hash password failed", zap.Error(err))
		return fmt.Errorf("hash password failed: %w", err)
	}
	u.HashedPassword = string(hash)
	u.Password = ""
	u.CreatedAt = time.Now().UTC()
	u.IsActive = true

	if err := s.store.Store(ctx, u); err != nil {
		s.logger.Error("store user failed", zap.Error(err))
		return err
	}

	s.logger.Info("user created", zap.String("national_id", string(u.NationalID)))
	return nil
}

func (s *user) GetAllUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	users, err := s.store.FindAllUsers(ctx, limit, offset)
	if err != nil {
		s.logger.Error("get all users failed", zap.Error(err))
		return nil, err
	}
	return users, nil
}

func (s *user) GetUserByNationalID(ctx context.Context, id domain.NationalID) (domain.User, error) {
	if err := id.Validate(); err != nil {
		return domain.User{}, err
	}
	u, err := s.store.FindUserByNationalID(ctx, id)
	if err != nil {
		s.logger.Warn("get user by national id failed", zap.String("national_id", string(id)), zap.Error(err))
		return domain.User{}, err
	}
	return u, nil
}

func (s *user) GetUserByUsername(ctx context.Context, username domain.Username) (domain.User, error) {
	if err := username.Validate(); err != nil {
		return domain.User{}, err
	}
	u, err := s.store.FindUserByUsername(ctx, username)
	if err != nil {
		s.logger.Warn("get user by username failed", zap.String("username", string(username)), zap.Error(err))
		return domain.User{}, err
	}
	return u, nil
}

func (s *user) GetUsersByRole(ctx context.Context, role string) ([]domain.User, error) {
	if role == "" {
		return nil, fmt.Errorf("role cannot be empty")
	}
	users, err := s.store.FindUsersByRole(ctx, role)
	if err != nil {
		s.logger.Warn("get users by role failed", zap.String("role", role), zap.Error(err))
		return nil, err
	}
	return users, nil
}

func (s *user) GetUserRole(ctx context.Context, id domain.NationalID) (domain.User, error) {
	if err := id.Validate(); err != nil {
		return domain.User{}, err
	}
	u, err := s.store.FindUserRole(ctx, id)
	if err != nil {
		s.logger.Warn("get user role failed", zap.String("national_id", string(id)), zap.Error(err))
		return domain.User{}, err
	}
	return u, nil
}

func (s *user) GetUserByMobileNumber(ctx context.Context, m domain.MobileNumber) (domain.User, error) {
	if err := m.Validate(); err != nil {
		return domain.User{}, err
	}
	u, err := s.store.FindUserByMobileNumber(ctx, m)
	if err != nil {
		s.logger.Warn("get user by mobile failed", zap.String("mobile", string(m)), zap.Error(err))
		return domain.User{}, err
	}
	return u, nil
}

func (s *user) Update(ctx context.Context, u domain.User, id domain.NationalID) error {
	if err := id.Validate(); err != nil {
		return err
	}
	if u.Email != "" {
		if err := u.Email.Validate(); err != nil {
			return err
		}
	}
	if u.UserName != "" {
		if err := u.UserName.Validate(); err != nil {
			return err
		}
	}
	if u.MobileNumber != "" {
		if err := u.MobileNumber.Validate(); err != nil {
			return err
		}
	}
	if u.Password != "" {
		hash, err := u.Password.Hash()
		if err != nil {
			s.logger.Error("hash password failed", zap.Error(err))
			return err
		}
		u.HashedPassword = string(hash)
		u.Password = ""
	}

	if err := s.store.Update(ctx, u, id); err != nil {
		s.logger.Error("update user failed", zap.String("national_id", string(id)), zap.Error(err))
		return err
	}
	return nil
}

func (s *user) Delete(ctx context.Context, id domain.NationalID) error {
	if err := id.Validate(); err != nil {
		return err
	}
	if err := s.store.Delete(ctx, id); err != nil {
		s.logger.Error("delete user failed", zap.String("national_id", string(id)), zap.Error(err))
		return err
	}
	return nil
}
