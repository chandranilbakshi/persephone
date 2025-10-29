package cmd

import (
	"Persephone/internal/purrCommands"
	"Persephone/internal/utils"
	"fmt"
	"github.com/spf13/cobra"
)

// Define the commit subcommand
var commitCmd = &cobra.Command{
	Use:   "commit -m [message]",
	Short: "Record changes to the repository",
	Run: func(cmd *cobra.Command, args []string) {
		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			fmt.Println("Error: commit message is required. Use -m \"message\"")
			return
		}

		// Check if user.name and user.email are configured
		userName, err := utils.ReadExistingConfig("user.name")
		if err != nil || userName == "" {
			fmt.Println("Error: user.name is not set.")
			fmt.Println("Please configure it using: purr config user.name \"Your Name\"")
			return
		}

		userEmail, err := purrCommands.ReadExistingConfig("user.email")
		if err != nil || userEmail == "" {
			fmt.Println("Error: user.email is not set.")
			fmt.Println("Please configure it using: purr config user.email \"your.email@example.com\"")
			return
		}

		// Now call CommitPurrFiles with the required parameters
		err = purrCommands.CommitPurrFiles(".", message, userName, userEmail)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("Changes committed successfully")
	},
}

func init() {
	// Add the -m flag for commit message
	commitCmd.Flags().StringP("message", "m", "", "Commit message")
	rootCmd.AddCommand(commitCmd)
}
