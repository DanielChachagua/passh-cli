package main

import (
	"fmt"
	"os"

	"ssh-cli/internal/adapters/api/http"
	"ssh-cli/internal/adapters/cli/cobra"
	"ssh-cli/internal/adapters/config/file"
	"ssh-cli/internal/adapters/ssh/crypto"
	"ssh-cli/internal/app"

	"github.com/spf13/viper"
)

func main() {
	// Configure Viper to check environment variables (e.g. API_URL)
	viper.SetDefault("API_URL", "http://localhost:8080")
	viper.AutomaticEnv()

	apiURL := viper.GetString("API_URL")

	// Instantiate Output Adapters
	configStorage, err := file.NewFileConfigStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize local config storage: %v\n", err)
		os.Exit(1)
	}

	apiScanner := http.NewHttpApiScanner(apiURL)
	sshDialer := crypto.NewSshCryptoDialer()

	// Instantiate Core Service (Input Port implementation)
	service := app.NewSshManagerService(apiScanner, configStorage, sshDialer)

	// Instantiate Input Adapter (Cobra CLI)
	cli := cobra.NewCLIAdapter(service)

	// Run Cobra Commands
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
