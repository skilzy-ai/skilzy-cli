
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/skilzy/skilzy-cli/utils"
	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish <path/to/package.skill>",
	Short: "Publish a skill to the Skilzy Registry",
	Long: `Publish a new or updated skill to the registry. Requires authentication via 'skilzy login'.

The package file should be a .skill or .zip file created with the 'skilzy package' command.`,
	Args: cobra.ExactArgs(1),
	Run:  runPublish,
}

func init() {
	rootCmd.AddCommand(publishCmd)
}

func runPublish(cmd *cobra.Command, args []string) {
	packagePath := args[0]

	fmt.Printf("📦 Publishing skill from: %s\n", packagePath)

	// Validate package file exists
	absPath, err := filepath.Abs(packagePath)
	if err != nil {
		fmt.Printf("✗ Invalid path: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("✗ Skill package not found at '%s'\n", absPath)
		os.Exit(1)
	}

	// Load API key
	apiKey, err := utils.LoadAPIKey()
	if err != nil {
		fmt.Printf("✗ Failed to load API key: %v\n", err)
		os.Exit(1)
	}

	if apiKey == "" {
		fmt.Println("✗ You must be logged in to publish a skill.")
		fmt.Println("  Please run 'skilzy login' first.")
		os.Exit(1)
	}

	// Create API client
	client := utils.NewSkilzyClient(apiKey)

	// Publish the skill
	fmt.Println("\n📤 Uploading skill package...")
	response, err := client.PublishSkill(absPath)
	if err != nil {
		fmt.Printf("\n✗ Failed to publish skill: %v\n", err)
		os.Exit(1)
	}

	// Show success message
	fmt.Println("\n✓ Publish request successful!")
	fmt.Printf("  - Skill: %s\n", response.Skill)
	fmt.Printf("  - Version: %s\n", response.Version)
	fmt.Printf("  - Status: %s\n", response.Status)

	if response.Status == "pending_review" {
		fmt.Println("\nℹ️  Your skill is now pending review. You'll be notified when it's approved.")
	}
}
