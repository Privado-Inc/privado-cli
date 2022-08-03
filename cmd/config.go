package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set config for Privado CLI",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
