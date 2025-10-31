
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/skilzy/skilzy-cli/utils"
	"github.com/spf13/cobra"
)

var (
	searchAuthor   string
	searchKeywords string
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for skills in the Skilzy Registry",
	Long: `Search for skills in the registry. Useful for discovering available skills
and checking if a skill name is already taken before creating one.

Examples:
  skilzy search "pdf"
  skilzy search "automation" --author skilzy-admin
  skilzy search "data" --keywords csv,excel`,
	Args: cobra.ExactArgs(1),
	Run:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringVar(&searchAuthor, "author", "", "Filter by author's username")
	searchCmd.Flags().StringVar(&searchKeywords, "keywords", "", "Comma-separated keywords to filter by")
}

func runSearch(cmd *cobra.Command, args []string) {
	query := args[0]

	fmt.Printf("🔍 Searching for '%s'...\n\n", query)

	// Parse keywords
	var keywords []string
	if searchKeywords != "" {
		keywords = strings.Split(searchKeywords, ",")
		for i := range keywords {
			keywords[i] = strings.TrimSpace(keywords[i])
		}
	}

	// Create client (no API key needed for search)
	client := utils.NewSkilzyClient("")

	// Search for skills
	results, err := client.SearchSkills(query, searchAuthor, keywords)
	if err != nil {
		fmt.Printf("✗ Search failed: %v\n", err)
		os.Exit(1)
	}

	// Display results
	if results.Total == 0 {
		fmt.Println("No skills found matching your criteria.")
		return
	}

	fmt.Printf("Found %d skill(s):\n\n", results.Total)

	// Print table header
	fmt.Printf("%-30s %-20s %-15s %s\n", "NAME", "AUTHOR", "VERSION", "DESCRIPTION")
	fmt.Println(strings.Repeat("-", 100))

	// Print results
	for _, skill := range results.Data {
		name := skill.Author + "/" + skill.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}
		author := skill.Author
		if len(author) > 20 {
			author = author[:17] + "..."
		}
		desc := skill.Description
		if len(desc) > 38 {
			desc = desc[:35] + "..."
		}
		fmt.Printf("%-30s %-20s %-15s %s\n", name, author, skill.LatestVersion, desc)
	}
}
