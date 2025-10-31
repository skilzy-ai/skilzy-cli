
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/skilzy/skilzy-cli/utils"
	"github.com/spf13/cobra"
)

var apiKeyFlag string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Skilzy Registry",
	Long:  `Save your API key for authenticated operations like publishing skills.`,
	Run:   runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "Your Skilzy API key (will prompt if not provided)")
}

func runLogin(cmd *cobra.Command, args []string) {
	var apiKey string

	// Get API key from flag or prompt
	if apiKeyFlag != "" {
		apiKey = apiKeyFlag
	} else {
		// Interactive prompt
		prompt := &survey.Password{
			Message: "Please enter your Skilzy API key:",
		}
		if err := survey.AskOne(prompt, &apiKey); err != nil {
			fmt.Printf("✗ Error reading API key: %v\n", err)
			os.Exit(1)
		}
	}

	// Trim whitespace
	apiKey = strings.TrimSpace(apiKey)

	// Validate API key
	if len(apiKey) < 10 {
		fmt.Println("✗ Invalid API key: must be at least 10 characters")
		os.Exit(1)
	}

	// Save API key
	if err := utils.SaveAPIKey(apiKey); err != nil {
		fmt.Printf("✗ Failed to save API key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ API key saved successfully")
	
	// Show where it was saved
	configPath, _ := utils.GetConfigPath()
	fmt.Printf("  Saved to: %s\n", configPath)
}