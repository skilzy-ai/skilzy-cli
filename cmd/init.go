
package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/skilzy/skilzy-cli/scaffold"
	"github.com/skilzy/skilzy-cli/utils"
	"github.com/spf13/cobra"
)

var yesFlag bool

var initCmd = &cobra.Command{
	Use:   "init [skill-name]",
	Short: "Initialize a new skill in the current directory",
	Long:  `Creates a new skill directory with a valid 'skill.json' manifest.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip interactive prompts and use default values")
}

func runInit(cmd *cobra.Command, args []string) {
	data := scaffold.SkillData{}
	isInteractive := !yesFlag && len(args) == 0

	if isInteractive {
		 if err := runInteractiveSurvey(&data); err != nil {
			fmt.Printf("✗ Aborted. %v\n", err)
			os.Exit(1)
		}
	} else {
		skillName := "my-new-skill"
		if len(args) > 0 {
			skillName = args[0]
			if err := validateSkillName(skillName); err != nil {
				fmt.Printf("✗ Invalid skill name: %v\n", err)
				os.Exit(1)
			}
		}
		fmt.Println("Running in non-interactive mode...")
		populateDefaultData(&data, skillName)
	}

	if err := scaffold.Create(data); err != nil {
		fmt.Printf("✗ Error creating skill: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✨ Skill initialized successfully!")
	fmt.Println("Next steps:")
	fmt.Printf("1. Edit %s/README.md for the registry details page and SKILL.md for AI agent instructions.\n", data.Name)
	fmt.Printf("2. Run 'cd %s && skilzy validate' to check your work.\n", data.Name)
}

func runInteractiveSurvey(data *scaffold.SkillData) error {
	defaultAuthor := utils.GetGitUserName()
	answers := struct {
		Name          string
		Description   string
		Author        string
		License       string
		RepositoryURL string
		Keywords      string
	}{}

	// Required questions
	var requiredQs = []*survey.Question{
		{Name: "name", Prompt: &survey.Input{Message: "Skill Name:", Help: "A unique, hyphen-case identifier."}, Validate: func(val interface{}) error { return validateSkillName(val.(string)) }},
		{Name: "description", Prompt: &survey.Input{Message: "Description:"}, Validate: survey.MinLength(20)},
		{Name: "author", Prompt: &survey.Input{Message: "Author:", Default: defaultAuthor}, Validate: survey.Required},
		{Name: "license", Prompt: &survey.Select{Message: "License:", Options: []string{"MIT", "Apache-2.0", "GPL-3.0", "BSD-3-Clause"}, Default: "MIT"}},
	}

	fmt.Println("This utility will walk you through creating a new skill.")
	if err := survey.Ask(requiredQs, &answers); err != nil {
		return err
	}

	// Optional questions - ask individually
	survey.AskOne(&survey.Input{Message: "GitHub Repository URL (optional):"}, &answers.RepositoryURL)
	survey.AskOne(&survey.Input{Message: "Keywords (comma-separated, optional):"}, &answers.Keywords)

	data.Name = answers.Name
	data.Description = answers.Description
	data.Author = answers.Author
	data.License = answers.License
	data.Version = "0.1.0"
	data.Entrypoint = "README.md"
	data.Icon = "assets/icon.svg"
	data.LicenseFile = "LICENSE"
	if answers.RepositoryURL != "" {
		data.Repository = &scaffold.Repository{Type: "git", URL: answers.RepositoryURL}
	}
	data.Runtime.Type = "python"
	data.Runtime.Version = ">=3.9"
	if answers.Keywords != "" {
		data.Keywords = regexp.MustCompile(`[\s,]+`).Split(strings.TrimSpace(answers.Keywords), -1)
	}

	return nil
}

func populateDefaultData(data *scaffold.SkillData, skillName string) {
	data.Name = skillName
	data.Version = "0.1.0"
	data.Description = "A new Skilzy skill. Please provide a detailed description of its capabilities."
	data.Author = utils.GetGitUserName()
	if data.Author == "" {
		data.Author = "Author Name"
	}
	data.License = "MIT"
	data.LicenseFile = "LICENSE"
	data.Entrypoint = "README.md"
	data.Icon = "assets/icon.svg"
	data.Runtime.Type = "python"
	data.Runtime.Version = ">=3.9"
	data.Keywords = []string{}
	data.Dependencies = &scaffold.Dependencies{
		System: []string{},
		Python: []string{},
		Skills: []string{},
	}
}

func validateSkillName(name string) error {
	if name == "" {
		return fmt.Errorf("skill name cannot be empty")
	}
	re := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	if !re.MatchString(name) {
		return fmt.Errorf("must be a hyphen-case identifier (e.g., 'my-skill')")
	}
	if len(name) > 40 {
		return fmt.Errorf("cannot be longer than 40 characters")
	}
	return nil
}
