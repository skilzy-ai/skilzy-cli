
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/skilzy/skilzy-cli/schema"
	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a skill manifest against the official schema",
	Long: `The validate command checks the skill in the current directory, ensuring
the skill.json is valid and all file references are correct.`,
	Run:  runValidate,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

// runValidate is the function executed by the 'validate' command.
func runValidate(cmd *cobra.Command, args []string) {
	fmt.Println("🔍 Running skill validation...")
	
	skillDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	validationErrors := doValidation(skillDir)

	if len(validationErrors) > 0 {
		fmt.Println("\n❌ Validation failed. Please fix the following issues:")
		for _, e := range validationErrors {
			fmt.Println(e)
		}
		os.Exit(1)
	}

	fmt.Println("\n✨ Skill is valid!")
}

// doValidation contains the core validation logic, designed to be reusable by other commands.
// It returns a slice of error strings if validation fails, or an empty slice if it succeeds.
func doValidation(skillDir string) []string {
	var allErrors []string
	manifestPath := filepath.Join(skillDir, "skill.json")

	// --- Pre-check: skill.json must exist ---
	manifestContent, err := os.ReadFile(manifestPath)
	if err != nil {
		return append(allErrors, "Validation failed: skill.json not found in the current directory. Ensure you are in a valid skill directory.")
	}
	
	// --- Schema Validation ---
	schemaLoader := gojsonschema.NewStringLoader(schema.SkillSchemaContent)
	manifestLoader := gojsonschema.NewStringLoader(string(manifestContent))

	result, err := gojsonschema.Validate(schemaLoader, manifestLoader)
	if err != nil {
		return append(allErrors, fmt.Sprintf("Error during validation: %v", err))
	}
	if !result.Valid() {
		allErrors = append(allErrors, "Schema validation failed with the following errors:")
		for _, desc := range result.Errors() {
			allErrors = append(allErrors, fmt.Sprintf("  - %s", desc))
		}
	} else {
		fmt.Println("✅ Schema validation successful.")
	}

	// --- Filesystem Checks ---
	fsErrors := performFileSystemChecks(skillDir, manifestPath)
	if len(fsErrors) > 0 {
		if len(allErrors) == 0 {
			allErrors = append(allErrors, "Filesystem checks failed with the following issues:")
		}
		for _, e := range fsErrors {
			allErrors = append(allErrors, fmt.Sprintf("  - %s", e))
		}
	} else {
		fmt.Println("✅ Filesystem checks successful.")
	}

	return allErrors
}

// performFileSystemChecks ensures files declared in the manifest exist.
func performFileSystemChecks(skillDir, manifestPath string) []string {
	var errors []string
	content, _ := os.ReadFile(manifestPath)
	var data struct {
		Name        string `json:"name"`
		Icon        string `json:"icon"`
		LicenseFile string `json:"licenseFile"`
		Entrypoint  string `json:"entrypoint"`
	}
	json.Unmarshal(content, &data)

	// Check directory name matches manifest name
	dirName := filepath.Base(skillDir)
	if dirName != data.Name {
		errors = append(errors, fmt.Sprintf("Directory name ('%s') does not match 'name' in skill.json ('%s').", dirName, data.Name))
	}
	
	// Check that declared files exist
	filesToCheck := []struct{path, fieldName string}{
		{data.Icon, "icon"}, {data.LicenseFile, "licenseFile"}, {data.Entrypoint, "entrypoint"},
	}
	for _, fileCheck := range filesToCheck {
		if fileCheck.path != "" {
			fullPath := filepath.Join(skillDir, fileCheck.path)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("File '%s' declared in '%s' field does not exist.", fileCheck.path, fileCheck.fieldName))
			}
		}
	}
	return errors
}
