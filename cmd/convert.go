
package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/skilzy/skilzy-cli/scaffold"
	"github.com/skilzy/skilzy-cli/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type FrontMatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

var convertCmd = &cobra.Command{
	Use:   "convert [path-to-skill.zip]",
	Short: "Convert a skill from another format to the Skilzy standard",
	Long:  `This command inspects an existing skill archive and guides you through generating a valid skill.json manifest.`,
	Args:  cobra.ExactArgs(1),
	Run:   runConvert,
}

func init() {
	rootCmd.AddCommand(convertCmd)
}

func runConvert(cmd *cobra.Command, args []string) {
	sourceZipPath := args[0]
	fmt.Printf("🔍 Analyzing skill package: %s\n", sourceZipPath)

	tempDir, err := os.MkdirTemp("", "skilzy-convert-*")
	if err != nil {
		fmt.Printf("✗ Failed to create temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	if err := unzipSource(sourceZipPath, tempDir); err != nil {
		fmt.Printf("✗ Failed to unzip source file: %v\n", err)
		os.Exit(1)
	}

	sourceSkillDir, frontMatter, err := analyzeSource(tempDir)
	if err != nil {
		fmt.Printf("✗ Analysis failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Analysis complete. Found the following metadata:")
	fmt.Printf("   Name: %s\n", frontMatter.Name)
	fmt.Printf("   Description: %s\n", frontMatter.Description)

	finalSkillData := scaffold.SkillData{}
	if err := runConversionSurvey(&finalSkillData, *frontMatter); err != nil {
		fmt.Printf("✗ Aborted. %v\n", err)
		os.Exit(1)
	}

	if err := generateConvertedSkill(&finalSkillData, sourceSkillDir); err != nil {
		fmt.Printf("✗ Failed to generate converted skill: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✨ Successfully converted and created skill '%s'.\n", finalSkillData.Name)
	fmt.Println("   Next, run 'cd '" + finalSkillData.Name + "' && skilzy validate' to confirm.")
}

func runConversionSurvey(data *scaffold.SkillData, fm FrontMatter) error {
	defaultAuthor := utils.GetGitUserName()
	answers := struct {
		Name, Description, Author, License, RepositoryURL, Keywords string
	}{}

	// Required questions
	var requiredQs = []*survey.Question{
		{Name: "name", Prompt: &survey.Input{Message: "Skill Name:", Default: fm.Name}, Validate: func(val interface{}) error { return validateSkillName(val.(string)) }},
		{Name: "description", Prompt: &survey.Input{Message: "Description:", Default: fm.Description}, Validate: survey.MinLength(20)},
		{Name: "author", Prompt: &survey.Input{Message: "Author:", Default: defaultAuthor}, Validate: survey.Required},
		{Name: "license", Prompt: &survey.Select{Message: "License:", Options: []string{"MIT", "Apache-2.0", "GPL-3.0"}, Default: "MIT"}},
	}

	fmt.Println("\nThis utility will convert the skill to the Skilzy format.")
	if err := survey.Ask(requiredQs, &answers); err != nil {
		return err
	}

	// Optional questions
	survey.AskOne(&survey.Input{Message: "GitHub Repository URL (optional):"}, &answers.RepositoryURL)
	survey.AskOne(&survey.Input{Message: "Keywords (comma-separated, optional):"}, &answers.Keywords)

	data.Name = answers.Name
	data.Description = answers.Description
	data.Author = answers.Author
	data.License = answers.License
	data.Version = "1.0.0"
	data.Entrypoint = "README.md"
	data.Icon = "assets/icon.svg"
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

func generateConvertedSkill(data *scaffold.SkillData, sourceDir string) error {
	destDir := data.Name
	if _, err := os.Stat(destDir); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", destDir)
	}
	if err := os.Mkdir(destDir, 0755); err != nil {
		return err
	}

	// Copy all files
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(sourceDir, path)
		destPath := filepath.Join(destDir, relPath)
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// If this is SKILL.md, strip frontmatter and keep it as SKILL.md
		if info.Name() == "SKILL.md" {
			if parts := strings.SplitN(string(content), "---", 3); len(parts) >= 3 {
				content = []byte(parts[2])
			}
			// Also create README.md from the SKILL.md content
			readmeContent := fmt.Sprintf("# %s\n\n## Overview\n\n%s\n\n## Description\n\n[Add detailed description here - this will be shown on the Skilzy registry]\n\n---\n\n**Note:** This README is displayed on the Skilzy registry details page. The actual AI agent instructions are in SKILL.md.\n",
				data.Name, data.Description)
			readmePath := filepath.Join(destDir, "README.md")
			if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
				return err
			}
		}

		return os.WriteFile(destPath, content, info.Mode())
	})
	if err != nil {
		return fmt.Errorf("failed to copy skill files: %w", err)
	}

	// Check for a license file and set the manifest field.
	foundLicenseFile := findLicenseFileInDir(destDir)
	if foundLicenseFile != "" {
		data.LicenseFile = foundLicenseFile
	} else {
		defaultLicensePath := filepath.Join(destDir, "LICENSE")
		os.WriteFile(defaultLicensePath, []byte("License content to be added here."), 0644)
		data.LicenseFile = "LICENSE"
	}

	// Create default assets/icon.svg if it wasn't copied.
	iconPath := filepath.Join(destDir, data.Icon)
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		assetsDir := filepath.Dir(iconPath)
		os.MkdirAll(assetsDir, 0755)
		iconContent := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 24 24"><path d="M12 2L2 7v10l10 5 10-5V7L12 2zm0 2.23L19.77 7 12 11.77 4.23 7 12 4.23zM3 8.5l9 5.06v9.44L3 17.5V8.5zm18 0v9l-9 5.06v-9.44L21 8.5z"/></svg>`
		if err := os.WriteFile(iconPath, []byte(iconContent), 0644); err != nil {
			return fmt.Errorf("failed to create default icon: %w", err)
		}
		fmt.Println("✅ Created default assets/icon.svg as it was missing from source.")
	}

	// Write the final manifest
	if err := scaffold.WriteManifest(destDir, *data); err != nil {
		return err
	}

	return nil
}

func analyzeSource(searchDir string) (string, *FrontMatter, error) {
	var skillMDPath string
	filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.Name() == "SKILL.md" {
			skillMDPath = path
			return filepath.SkipDir
		}
		return nil
	})
	if skillMDPath == "" {
		return "", nil, fmt.Errorf("could not find SKILL.md in the archive")
	}
	content, err := os.ReadFile(skillMDPath)
	if err != nil {
		return "", nil, fmt.Errorf("could not read SKILL.md: %w", err)
	}
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return "", nil, fmt.Errorf("SKILL.md has no YAML frontmatter")
	}
	var fm FrontMatter
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return "", nil, err
	}
	if fm.Name == "" || fm.Description == "" {
		return "", nil, fmt.Errorf("YAML missing 'name' or 'description'")
	}
	return filepath.Dir(skillMDPath), &fm, nil
}

func findLicenseFileInDir(dir string) string {
	for _, name := range []string{"LICENSE.txt", "LICENSE.md", "LICENSE"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return name
		}
	}
	return ""
}

func unzipSource(source, dest string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}
		destFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		_, err = io.Copy(destFile, rc)
		destFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
