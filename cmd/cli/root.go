package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgPath   string
	portFlag  string
	bodyLimit int64
	swagger   string

	rootCmd = &cobra.Command{
		Use:   "user-mgmt",
		Short: "User Management service CLI",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "configs/config.yaml", "Path to config file")
	rootCmd.PersistentFlags().StringVar(&portFlag, "port", "", "Override HTTP listen port (optional)")
	rootCmd.PersistentFlags().Int64Var(&bodyLimit, "body-limit-bytes", 1048576, "Maximum request body size in bytes")
	rootCmd.PersistentFlags().StringVar(&swagger, "swagger", "docs/swagger.json", "Path to static swagger.json")
}

func CfgPath() string { return cfgPath }

func PortOverride() string { return portFlag }

func BodyLimit() int64 { return bodyLimit }

func SwaggerPath() string { return swagger }
