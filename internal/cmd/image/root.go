package image

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(createdCmd)
	RootCmd.AddCommand(labelsCmd)
	RootCmd.PersistentFlags().Bool("insecure", false, "Allow communication with an insecure registry")
}

var RootCmd = &cobra.Command{
	Use:   "image",
	Short: "Image related commands",
}
