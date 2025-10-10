package main

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/user-management/configs"
	"github.com/mohammadgholizadeh/user-management/internal/domain"
	"github.com/mohammadgholizadeh/user-management/internal/storage"
	db "github.com/mohammadgholizadeh/user-management/pkg/database"
	logger "github.com/mohammadgholizadeh/user-management/pkg/logger"
)

var doAdmin bool

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed initial data into the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Logger
		logger := logger.NewLogger()
		defer func() { _ = logger.Sync() }()
		zap.ReplaceGlobals(logger)

		cfg := configs.LoadConfig(CfgPath())

		pool, err := db.NewPostgresPool(cfg.DB, logger)
		if err != nil {
			logger.Fatal("postgres connection failed", zap.Error(err))
		}
		defer pool.Close()

		adminStorage := storage.NewAdminStorage(pool, logger)
		userStorage := storage.NewUserStore(pool, logger)

		if doAdmin {
			u := domain.User{
				NationalID:   "2526738261",
				FirstName:    "super",
				LastName:     "admin",
				Password:     "12345678",
				UserName:     "superadmin",
				Role:         "admin",
				MobileNumber: "09123456789",
				IsActive:     true,
			}

			hashed, _ := u.Password.Hash()
			u.HashedPassword = string(hashed)

			err := userStorage.Store(context.Background(), u)
			if err != nil {
				return err
			}

			a := domain.Admin{UserNationalID: u.NationalID}
			err = adminStorage.Store(context.Background(), a)
			if err != nil {
				return err
			}

			logger.Info("Admin user seeded successfully")
		}

		return nil
	},
}

func init() {
	seedCmd.Flags().BoolVar(&doAdmin, "doAdmin", false, "seed admin user")
	rootCmd.AddCommand(seedCmd)
}
