
package scaffold

import (
    "bytes"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

type SkillData struct {
    Name         string        `json:"name"`
    Version      string        `json:"version"`
    Description  string        `json:"description"`
    Author       string        `json:"author"`
    License      string        `json:"license"`
    LicenseFile  string        `json:"licenseFile"`
    Repository   *Repository   `json:"repository,omitempty"`
    Icon         string        `json:"icon"`
    Entrypoint   string        `json:"entrypoint"`
    Runtime      Runtime       `json:"runtime"`
    Dependencies *Dependencies `json:"dependencies,omitempty"`
    Permissions  *Permissions  `json:"permissions,omitempty"`
    Keywords     []string      `json:"keywords,omitempty"`
}

type Repository struct {
    Type string `json:"type"`
    URL  string `json:"url"`
}

type Runtime struct {
    Type    string `json:"type"`
    Version string `json:"version"`
}

type Dependencies struct {
    System []string `json:"system,omitempty"`
    Python []string `json:"python,omitempty"`
    Skills []string `json:"skills,omitempty"`
}

type Permissions struct {
    // Empty for now, can be expanded later
}

// Create generates the directory and skill.json for a new skill from scratch.
func Create(data SkillData) error {
    skillDir := data.Name
    fmt.Printf("🚀 Initializing skill: %s...\n", skillDir)

    if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
        return fmt.Errorf("directory '%s' already exists", skillDir)
    }
    if err := os.Mkdir(skillDir, 0755); err != nil {
        return fmt.Errorf("failed to create skill directory: %w", err)
    }
    fmt.Printf("✅ Created skill directory: ./%s\n", skillDir)

    if err := WriteManifest(skillDir, data); err != nil {
        return err
    }

    // Create README.md (entrypoint - shown on registry)
    readmePath := filepath.Join(skillDir, data.Entrypoint)
    skillTitle := strings.ToTitle(strings.ReplaceAll(data.Name, "-", " "))
    readmeContent := fmt.Sprintf("# %s\n\n## Overview\n\n%s\n\n## Features\n\n- Feature 1: [Describe key capability]\n- Feature 2: [Describe another capability]\n- Feature 3: [Describe additional functionality]\n\n## Usage\n\n[Provide examples of how to use this skill]\n\n## Requirements\n\n- Python >= 3.9\n- [List any other dependencies]\n\n## Installation\n\nThis skill is available through the Skilzy Registry. Install it using:\n\n```bash\nskilzy install your-username/%s\n```\n\n## License\n\n%s\n\n## Author\n\n%s\n\n---\n\n**Note:** This README is displayed on the Skilzy registry details page. The actual AI agent instructions are in SKILL.md.\n", 
        skillTitle, data.Description, data.Name, data.License, data.Author)
    if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
        return fmt.Errorf("failed to write README.md: %w", err)
    }
    fmt.Printf("✅ Created %s\n", readmePath)

    // Create SKILL.md (AI agent instructions)
    skillMDPath := filepath.Join(skillDir, "SKILL.md")
    skillMDContent := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n# %s\n\n## Overview\n\nThis skill is designed to...\n\n## When to Use This Skill\n\nThis skill should be used when...\n\n## Instructions\n\nRun 'skilzy validate' when ready\n\n## Resources\n\n[Reference any scripts, references, or assets included with this skill]\n", 
        data.Name, data.Description, skillTitle)
    if err := os.WriteFile(skillMDPath, []byte(skillMDContent), 0644); err != nil {
        return fmt.Errorf("failed to write SKILL.md: %w", err)
    }
    fmt.Printf("✅ Created SKILL.md\n")

    licensePath := filepath.Join(skillDir, data.LicenseFile)
    if err := os.WriteFile(licensePath, []byte(fmt.Sprintf("This project is licensed under the %s license.", data.License)), 0644); err != nil {
        return fmt.Errorf("failed to write LICENSE: %w", err)
    }
    fmt.Printf("✅ Created %s\n", licensePath)

    subDirs := []string{"assets", "scripts", "reference"}
    for _, dir := range subDirs {
        fullPath := filepath.Join(skillDir, dir)
        if err := os.Mkdir(fullPath, 0755); err != nil {
            return fmt.Errorf("failed to create directory %s: %w", dir, err)
        }
        fmt.Printf("✅ Created directory: %s/\n", fullPath)
    }

    iconPath := filepath.Join(skillDir, data.Icon)
    iconContent := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><svg id=\"a\" xmlns=\"http://www.w3.org/2000/svg\" width=\"12.52mm\" height=\"12.09mm\" viewBox=\"0 0 35.48 34.27\"><rect x=\"0\" y=\"0\" width=\"35.48\" height=\"34.27\" style=\"fill:#755541;\"/><path d=\"M6.13,26.42v-2.85h9.88v2.85H6.13Z\" style=\"fill:#fff;\"/><path d=\"M24.09,23.93c-1.56,0-2.82-.25-3.78-.76-.96-.51-1.69-1.12-2.19-1.85-.5-.73-.82-1.44-.96-2.13l3.52-1.15c.16.63.51,1.19,1.03,1.69s1.32.74,2.37.74c.95,0,1.6-.15,1.96-.44.35-.29.53-.61.53-.96,0-.4-.21-.76-.64-1.06-.42-.3-1.28-.51-2.55-.61-1.68-.14-3.03-.59-4.04-1.34-1.01-.75-1.52-1.83-1.52-3.25v-.18c0-.99.27-1.83.82-2.52.55-.69,1.26-1.21,2.13-1.56s1.79-.53,2.76-.53c1.28,0,2.33.2,3.16.59.83.39,1.48.89,1.96,1.47.48.59.81,1.17,1.02,1.76l-3.46,1.4c-.18-.55-.5-.98-.94-1.29-.45-.31-1.02-.47-1.73-.47-.63,0-1.11.12-1.44.35-.33.23-.5.52-.5.87,0,.49.25.83.76,1.03.51.2,1.43.38,2.76.55,1.54.16,2.79.62,3.77,1.38.97.76,1.46,1.82,1.46,3.17v.18c0,1.48-.55,2.67-1.64,3.57-1.09.9-2.63,1.35-4.62,1.35Z\" style=\"fill:#fff;\"/></svg>"
    if err := os.WriteFile(iconPath, []byte(iconContent), 0644); err != nil {
        return fmt.Errorf("failed to write icon.svg: %w", err)
    }
    fmt.Printf("✅ Created %s\n", iconPath)

    return nil
}

// WriteManifest is a helper to only write the skill.json file.
func WriteManifest(skillDir string, data SkillData) error {
	manifestPath := filepath.Join(skillDir, "skill.json")
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to serialize skill.json: %w", err)
	}
	if err := os.WriteFile(manifestPath, buffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write skill.json: %w", err)
	}
	fmt.Printf("✅ Created %s\n", manifestPath)
	return nil
}
