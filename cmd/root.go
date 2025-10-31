
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "skilzy",
	Short: "A CLI for creating, managing, and publishing AI skills",
	Long: `Skilzy CLI is a command-line interface to the Skilzy.ai registry.
It helps you create, validate, package, and publish skills for AI agents.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
