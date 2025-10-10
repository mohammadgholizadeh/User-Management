package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mohammadgholizadeh/user-management/internal/domain"
	"github.com/mohammadgholizadeh/user-management/pkg/jwt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type authentication struct {
	userService     domain.UserService
	logger          *zap.Logger
	cache           *redis.Client
	expirationHours time.Duration
	jwtManager      *jwt.Manager
}

func NewAuthenticationService(
	userService domain.UserService,
	logger *zap.Logger,
	cache *redis.Client,
	expirationHours time.Duration,
	jwtManager *jwt.Manager,
) domain.AuthenticationService {
	return &authentication{
		userService:     userService,
		logger:          logger,
		cache:           cache,
		expirationHours: expirationHours,
		jwtManager:      jwtManager,
	}
}

type ctxKey string

const tokenCtxKey ctxKey = "auth_token"

func PutTokenInContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenCtxKey, token)
}

func (a *authentication) Login(ctx context.Context, dto domain.LoginRequest) (domain.LoginResponse, error) {
	if err := dto.Usrname.Validate(); err != nil {
		return domain.LoginResponse{}, fmt.Errorf("invalid username: %w", err)
	}
	if err := dto.Password.Validate(); err != nil {
		return domain.LoginResponse{}, fmt.Errorf("invalid password: %w", err)
	}

	user, err := a.userService.GetUserByUsername(ctx, dto.Usrname)
	if err != nil {
		a.logger.Warn("login: user not found", zap.String("username", string(dto.Usrname)), zap.Error(err))
		return domain.LoginResponse{}, fmt.Errorf("invalid credentials")
	}
	if !user.IsActive {
		a.logger.Info("login: inactive user", zap.String("username", string(dto.Usrname)))
		return domain.LoginResponse{}, fmt.Errorf("user is inactive")
	}
	if !dto.Password.CompareWithHashedPassword(user.HashedPassword) {
		a.logger.Info("login: wrong password", zap.String("username", string(dto.Usrname)))
		return domain.LoginResponse{}, fmt.Errorf("invalid credentials")
	}

	token, err := a.jwtManager.GenerateToken(string(user.UserName), user.Role, a.expirationHours)
	if err != nil {
		a.logger.Error("jwt generate failed", zap.Error(err))
		return domain.LoginResponse{}, fmt.Errorf("failed to generate token")
	}

	return domain.LoginResponse{Token: domain.AuthToken(token)}, nil
}

func (a *authentication) Register(ctx context.Context, dto domain.RegisterRequest) (domain.RegisterResponse, error) {
	if err := dto.NationalID.Validate(); err != nil {
		return domain.RegisterResponse{}, err
	}
	if err := dto.Username.Validate(); err != nil {
		return domain.RegisterResponse{}, err
	}
	if err := dto.Password.Validate(); err != nil {
		return domain.RegisterResponse{}, err
	}
	if err := dto.Email.Validate(); err != nil {
		return domain.RegisterResponse{}, err
	}
	if err := dto.MobileNumber.Validate(); err != nil {
		return domain.RegisterResponse{}, err
	}

	if u, err := a.userService.GetUserByNationalID(ctx, dto.NationalID); err == nil && u.NationalID != "" {
		a.logger.Info("register: duplicate national id", zap.String("national_id", string(dto.NationalID)))
		return domain.RegisterResponse{}, fmt.Errorf("user already exists")
	}
	if u, err := a.userService.GetUserByUsername(ctx, dto.Username); err == nil && u.UserName != "" {
		a.logger.Info("register: duplicate username", zap.String("username", string(dto.Username)))
		return domain.RegisterResponse{}, fmt.Errorf("username already taken")
	}

	user := domain.User{
		NationalID:   dto.NationalID,
		FirstName:    dto.FirstName,
		LastName:     dto.LastName,
		UserName:     dto.Username,
		Password:     dto.Password,
		MobileNumber: dto.MobileNumber,
		Email:        dto.Email,
		Gender:       dto.Gender,
		Role:         strings.TrimSpace(dto.Role),
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
	}

	if err := a.userService.Create(ctx, user); err != nil {
		a.logger.Error("register: create user failed", zap.Error(err))
		return domain.RegisterResponse{}, err
	}

	token, err := a.jwtManager.GenerateToken(string(user.UserName), user.Role, a.expirationHours)
	if err != nil {
		a.logger.Error("jwt generate failed", zap.Error(err))
		return domain.RegisterResponse{}, fmt.Errorf("failed to generate token")
	}

	return domain.RegisterResponse{User: user, Token: domain.AuthToken(token)}, nil
}

func (a *authentication) Logout(ctx context.Context) error {
	v := ctx.Value(tokenCtxKey)
	token, _ := v.(string)
	if token == "" {
		return nil
	}

	if a.cache != nil {
		if err := a.cache.Set(ctx, "bl:"+token, "1", 24*time.Hour).Err(); err != nil {
			a.logger.Error("redis set blacklist failed", zap.Error(err))
			return err
		}
	}
	return nil
}

func (a *authentication) SendResetPasswordToken(ctx context.Context, username domain.Username, nationalId domain.NationalID) error {
	if username == "" && nationalId == "" {
		return fmt.Errorf("username or national_id required")
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return err
	}
	token := hex.EncodeToString(b)

	key := "rp:"
	if username != "" {
		key += "u:" + string(username)
	}
	if nationalId != "" {
		key += ":n:" + string(nationalId)
	}
	if a.cache != nil {
		if err := a.cache.Set(ctx, key, token, 15*time.Minute).Err(); err != nil {
			a.logger.Error("redis set reset token failed", zap.Error(err))
			return err
		}
	}
	a.logger.Info("reset password token generated", zap.String("key", key))
	return nil
}

func (a *authentication) confirmResetPasswordToken(ctx context.Context, username domain.Username, nationalId domain.NationalID, token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("token required")
	}
	key := "rp:"
	if username != "" {
		key += "u:" + string(username)
	}
	if nationalId != "" {
		key += ":n:" + string(nationalId)
	}
	if a.cache == nil {
		return "", errors.New("cache is not initialized")
	}

	val, err := a.cache.Get(ctx, key).Result()
	if err != nil {
		a.logger.Error("redis get reset token failed", zap.Error(err))
		return "", err
	}
	if val != token {
		return "", fmt.Errorf("invalid token")
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	confirm := hex.EncodeToString(b)
	if err := a.cache.Set(ctx, "rpconf:"+key, confirm, 15*time.Minute).Err(); err != nil {
		a.logger.Error("redis set confirm token failed", zap.Error(err))
		return "", err
	}
	return confirm, nil
}

func (a *authentication) ResetPassword(ctx context.Context, username domain.Username, nationalId domain.NationalID, token string, newPassword string) error {
	if token == "" {
		return fmt.Errorf("token required")
	}
	confirmKey := "rpconf:rp:"
	if username != "" {
		confirmKey += "u:" + string(username)
	}
	if nationalId != "" {
		confirmKey += ":n:" + string(nationalId)
	}
	resetKey := "rp:"
	if username != "" {
		resetKey += "u:" + string(username)
	}
	if nationalId != "" {
		resetKey += ":n:" + string(nationalId)
	}
	if a.cache == nil {
		return errors.New("cache is not initialized")
	}

	val, err := a.cache.Get(ctx, confirmKey).Result()
	if err != nil {
		val, err = a.cache.Get(ctx, resetKey).Result()
		if err != nil {
			a.logger.Error("redis get reset token failed", zap.Error(err))
			return err
		}
		if val != token {
			return fmt.Errorf("invalid token")
		}
		_ = a.cache.Del(ctx, resetKey).Err()
	} else {
		if val != token {
			return fmt.Errorf("invalid token")
		}
		_ = a.cache.Del(ctx, confirmKey).Err()
	}

	if err := domain.Password(newPassword).Validate(); err != nil {
		return err
	}

	var u domain.User
	if username != "" {
		u, err = a.userService.GetUserByUsername(ctx, username)
	} else {
		u, err = a.userService.GetUserByNationalID(ctx, nationalId)
	}
	if err != nil {
		return err
	}

	hash, err := domain.Password(newPassword).Hash()
	if err != nil {
		return err
	}
	u.HashedPassword = string(hash)
	u.Password = ""

	if err := a.userService.Update(ctx, u, u.NationalID); err != nil {
		return err
	}

	_ = a.cache.Del(ctx, confirmKey).Err()
	_ = a.cache.Del(ctx, resetKey).Err()

	return nil
}
