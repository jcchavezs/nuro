package labels

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(ListCmd)
	RootCmd.PersistentFlags().Bool("insecure", false, "Allow communication with an insecure registry")
}

var RootCmd = &cobra.Command{
	Use:   "labels",
	Short: "Labels related commands",
}
