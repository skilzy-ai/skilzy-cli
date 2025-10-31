
package cmd

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var outputDir string
var outputName string

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Validate and package a skill into a distributable .skill file",
	Long: `This command first validates the skill in the current directory.
If the skill is valid, it bundles all its files into a compressed .skill archive
containing a single root folder, ready for distribution.`,
	Run:  runPackage,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(packageCmd)
	packageCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "dist", "Directory to save the packaged skill (relative to the project root)")
	packageCmd.Flags().StringVar(&outputName, "output-name", "", "Specify a custom name for the output .skill file")
}

func runPackage(cmd *cobra.Command, args []string) {
	skillDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("📦 Starting package process...")
	validationErrors := doValidation(skillDir)
	if len(validationErrors) > 0 {
		fmt.Println("\n❌ Validation failed. Cannot package an invalid skill.")
		fmt.Println("   Please fix the issues reported above and try again.")
		os.Exit(1)
	}
	fmt.Println("✨ Skill is valid, proceeding with packaging.")

	manifestPath := filepath.Join(skillDir, "skill.json")
	content, _ := os.ReadFile(manifestPath)
	var data struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	json.Unmarshal(content, &data)

	projectRoot := filepath.Dir(skillDir)
	finalOutputDir := filepath.Join(projectRoot, outputDir)
	if err := os.MkdirAll(finalOutputDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create output directory %s: %v\n", finalOutputDir, err)
		os.Exit(1)
	}

	archiveFileName := outputName
	if archiveFileName == "" {
		archiveFileName = fmt.Sprintf("%s-%s.skill", data.Name, data.Version)
	}
	archivePath := filepath.Join(finalOutputDir, archiveFileName)
	
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		fmt.Printf("❌ Failed to create archive file: %v\n", err)
		os.Exit(1)
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	err = filepath.Walk(skillDir, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }

		if info.IsDir() && path == finalOutputDir {
			return filepath.SkipDir
		}
		
		if path == skillDir { return nil }

		header, err := zip.FileInfoHeader(info)
		if err != nil { return err }
		
		relativePath, err := filepath.Rel(skillDir, path)
		if err != nil { return err }
		header.Name = filepath.Join(data.Name, relativePath)
		header.Name = filepath.ToSlash(header.Name)

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil { return err }

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil { return err }
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		}
		return nil
	})

	if err != nil {
		fmt.Printf("❌ Failed to add files to archive: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Successfully packaged skill to: %s\n", archivePath)
}
