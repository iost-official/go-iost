package iwallet

import (
	"github.com/spf13/cobra"
)

// systemCmd represents the system command.
var systemCmd = &cobra.Command{
	Use:     "system",
	Aliases: []string{"sys"},
	Short:   "Common system contract actions",
	Long:    `Common system contract actions`,
	Example: `  iwallet system producer-list
  iwallet sys producer-list
  iwallet sys plist`,
}

func init() {
	rootCmd.AddCommand(systemCmd)
}
