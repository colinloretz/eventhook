package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "eventhook",
	Short: "Webhook infrastructure runtime",
	Long:  "EventHook — Full webhook observability for every event in your app.",
}

func main() {
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(startCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
