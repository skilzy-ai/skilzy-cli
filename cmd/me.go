
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/skilzy/skilzy-cli/utils"
	"github.com/spf13/cobra"
)

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Manage your Skilzy account and published skills",
	Long:  `Commands for managing your account and viewing your published skills.`,
}

var meWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Validate the currently configured API key",
	Long:  `Validates your saved API key by testing it against the Skilzy API.`,
	 Run:   runMeWhoami,
}

var meSkillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "List all skills you have published",
	Long:  `Shows all skills you've published to the registry with their version and review status.`,
	Run:   runMeSkills,
}

func init() {
	rootCmd.AddCommand(meCmd)
	meCmd.AddCommand(meWhoamiCmd)
	meCmd.AddCommand(meSkillsCmd)
}

func runMeWhoami(cmd *cobra.Command, args []string) {
	// Load API key
	apiKey, err := utils.LoadAPIKey()
	if err != nil {
		fmt.Printf("✗ Failed to load API key: %v\n", err)
		os.Exit(1)
	}

	if apiKey == "" {
		fmt.Println("✗ No API key found.")
		fmt.Println("  Please run 'skilzy login' first.")
		os.Exit(1)
	}

	// Show key prefix
	keyPrefix := apiKey
	if len(apiKey) > 8 {
		keyPrefix = apiKey[:8] + "..."
	}
	fmt.Printf("Loaded API key prefix: %s\n", keyPrefix)

	// Validate with API
	fmt.Println("Attempting to validate key with the API...")

	client := utils.NewSkilzyClient(apiKey)
	_, err = client.GetMySkills()
	if err != nil {
		if strings.Contains(err.Error(), "authentication failed") {
			fmt.Println("\n✗ Validation failed: The API rejected this key (401 Unauthorized).")
			fmt.Println("  Please verify this key is correct or re-run 'skilzy login'.")
			os.Exit(1)
		}
		fmt.Printf("✗ Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Validation successful: The API accepted this key.")
}

func runMeSkills(cmd *cobra.Command, args []string) {
	// Load API key
	apiKey, err := utils.LoadAPIKey()
	if err != nil {
		fmt.Printf("✗ Failed to load API key: %v\n", err)
		os.Exit(1)
	}

	if apiKey == "" {
		fmt.Println("✗ You must be logged in.")
		fmt.Println("  Please run 'skilzy login' first.")
		os.Exit(1)
	}

	// Get published skills
	client := utils.NewSkilzyClient(apiKey)
	skills, err := client.GetMySkills()
	if err != nil {
		fmt.Printf("✗ Failed to retrieve skills: %v\n", err)
		os.Exit(1)
	}

	// Display results
	if len(skills) == 0 {
		fmt.Println("You have not published any skills yet.")
		return
	}

	fmt.Printf("You have published %d skill(s):\n\n", len(skills))

	// Print table header
	fmt.Printf("%-30s %-15s %-20s %s\n", "NAME", "LATEST VERSION", "STATUS", "PUBLISHED/TOTAL VERSIONS")
	fmt.Println(strings.Repeat("-", 90))

	// Print results
	for _, skill := range skills {
		name := skill.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		version := "N/A"
		status := "N/A"
		if skill.LatestVersion != nil {
			version = skill.LatestVersion.Version
			status = skill.LatestVersion.Status
		}

		versionInfo := fmt.Sprintf("%d/%d", skill.PublishedVersionCount, skill.TotalVersions)

		fmt.Printf("%-30s %-15s %-20s %s\n", name, version, status, versionInfo)
	}
}
