package cmd

import (
	"fmt"
	"os"

	// homedir "github.com/mitchellh/go-homedir"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "privado",
	Short: "Privado is a CLI tool that scans & monitors your repositories to build privacy, transparency reports & finds privacy issues",
	Long:  "Privado is a CLI tool that scans & monitors your repositories to build privacy, transparency reports & finds privacy issues. \nFind more at: https://github.com/Privado-Inc/privado",
}

func init() {

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		exit(fmt.Sprintln(err), true)
	}
}

func exit(msg string, error bool) {
	fmt.Println(msg)
	if error {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
