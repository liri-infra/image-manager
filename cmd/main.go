// SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/liri-infra/image-manager/internal/client"
	"github.com/liri-infra/image-manager/internal/logger"
	"github.com/liri-infra/image-manager/internal/server"
)

func genTokenCmd() *cobra.Command {
	var (
		configPath string
		verbose    bool
	)

	var cmd = &cobra.Command{
		Use:   "gentoken",
		Short: "Creates a new API token",
		Long:  "Generates a token that gives access to the API.",
		Run: func(cmd *cobra.Command, args []string) {
			// Toggle debug output
			logger.SetVerbose(verbose)

			// Validate arguments
			if len(configPath) == 0 {
				logger.Fatal("Path to configuration file is mandatory")
				return
			}

			// Open configuration file
			config, err := server.CreateConfig(configPath)
			if err != nil {
				logger.Fatalf("Cannot open configuration file: %v", err)
				return
			}

			// Generate token
			token, err := server.GenerateToken()
			if err != nil {
				logger.Fatalf("Failed to generate token: %v", err)
				return
			}

			// Save token to the configuration
			config.Tokens = append(config.Tokens, token)
			if err := config.Save(); err != nil {
				logger.Fatalf("Cannot save configuration file: %v", err)
				return
			}

			// Print token
			logger.Infof("Token: %s", token.Token)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "image-manager.yaml", "path to configuration file")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "more messages during the build")

	return cmd
}

func serverCmd() *cobra.Command {
	var (
		bindAddress string
		configPath  string
		storagePath string
		verbose     bool
	)

	var cmd = &cobra.Command{
		Use:   "server",
		Short: "Start the server",
		Run: func(cmd *cobra.Command, args []string) {
			// Toggle debug output
			logger.SetVerbose(verbose)

			// Open configuration file
			config, err := server.OpenConfig(configPath)
			if err != nil {
				logger.Fatalf("Cannot open configuration file: %v", err)
				return
			}

			// Overwrite storage path
			if storagePath != "" {
				config.StorageDir = storagePath
			}

			// We need a storage path
			if config.StorageDir == "" {
				logger.Fatal("Storage path is not configured")
				return
			}

			// Create channels directories
			for _, channel := range config.Channels {
				path := filepath.Join(config.StorageDir, channel.Path)
				if err := os.MkdirAll(path, 0755); err != nil {
					logger.Fatalf("Failed to create \"%s\": %v", path, err)
					return
				}
			}

			// Remove old images
			ticker := time.NewTicker(3600)
			go func() {
				for _ = range ticker.C {
					server.RemoveOldImages(config.StorageDir, config.Channels)
				}
			}()

			appState := &server.AppState{Config: config}
			if err := server.StartServer(bindAddress, appState); err != nil {
				logger.Fatal(err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "image-manager.yaml", "path to configuration file")
	cmd.Flags().StringVarP(&bindAddress, "address", "a", ":8080", "host name and port to bind")
	cmd.Flags().StringVarP(&storagePath, "path", "p", "", "override configured storage path")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "more messages during the build")

	return cmd
}

func clientCmd() *cobra.Command {
	var (
		url              string
		token            string
		channel          string
		isoFileName      string
		checksumFileName string
		verbose          bool
	)

	var cmd = &cobra.Command{
		Use:   "client",
		Short: "Upload images and to the repository",
		Run: func(cmd *cobra.Command, args []string) {
			// Toggle debug output
			logger.SetVerbose(verbose)

			// Check the token
			if token == "" {
				token = os.Getenv("IMAGE_MANAGER_TOKEN")
			}
			if token == "" {
				logger.Fatal("Token is mandatory")
				return
			}

			if channel == "" {
				logger.Fatal("Channel is mandatory")
				return
			}

			if isoFileName == "" {
				logger.Fatal("ISO file is mandatory")
				return
			}

			if checksumFileName == "" {
				logger.Fatal("Checksum file is mandatory")
				return
			}

			paths := []string{isoFileName, checksumFileName}

			if err := client.StartClient(url, token, channel, paths); err != nil {
				logger.Fatal(err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&url, "address", "a", "http://localhost:8080", "host name and port of the server")
	cmd.Flags().StringVarP(&token, "token", "t", "", "token to authenticate with the server")
	cmd.Flags().StringVarP(&channel, "channel", "c", "", "image channel name")
	cmd.Flags().StringVarP(&isoFileName, "iso", "", "", "ISO file to upload")
	cmd.Flags().StringVarP(&checksumFileName, "checksum", "", "", "Checksum file to upload")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "more messages during the build")

	return cmd
}

func init() {
	// Set logger flags
	log.SetFlags(0)
}

func main() {
	// Root command
	var rootCmd = &cobra.Command{
		Use:   "image-manager",
		Short: "Store images produced by a build server and manages them",
	}

	rootCmd.AddCommand(
		genTokenCmd(),
		serverCmd(),
		clientCmd(),
	)

	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
