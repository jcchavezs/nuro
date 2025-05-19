package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/jcchavezs/nuro/internal/auth"
	"github.com/jcchavezs/nuro/internal/cmd/created"
	"github.com/jcchavezs/nuro/internal/cmd/labels"
	"github.com/jcchavezs/nuro/internal/log"

	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var LevelIds = map[zapcore.Level][]string{
	zap.DebugLevel: {"debug"},
	zap.InfoLevel:  {"info"},
	zap.WarnLevel:  {"warn"},
	zap.ErrorLevel: {"error"},
}

var loglevel zapcore.Level = zapcore.ErrorLevel

func init() {
	RootCmd.PersistentFlags().Var(
		enumflag.New(&loglevel, "string", LevelIds, enumflag.EnumCaseInsensitive),
		"log-level",
		"Sets the log level",
	)

	// Check netrc information https://www.gnu.org/software/inetutils/manual/html_node/The-_002enetrc-file.html
	RootCmd.PersistentFlags().String("netrc-file", "", "Read .netrc from file location, has precedence over --netrc-stdin")
	RootCmd.PersistentFlags().Bool("netrc-stdin", false, "Read .netrc from stdin")

	RootCmd.AddCommand(created.RootCmd)
	RootCmd.AddCommand(labels.RootCmd)
}

var RootCmd = &cobra.Command{
	Use:   "nuro",
	Short: "Get more information about your docker images",
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		log.Init(loglevel, cmd.ErrOrStderr())

		if netRCFile, _ := cmd.Flags().GetString("netrc-file"); netRCFile != "" {
			if err := auth.LoadNetRCFile(cmd.Context(), netRCFile); err != nil {
				return fmt.Errorf("loading netrc file: %w", err)
			}
		} else if netRCFromStdin, _ := cmd.Flags().GetBool("netrc-stdin"); netRCFromStdin {
			if stdin, err := io.ReadAll(os.Stdin); err != nil {
				return fmt.Errorf("reading netrc from stdin: %w", err)
			} else {
				if err := auth.LoadNetRC(cmd.Context(), string(stdin)); err != nil {
					return fmt.Errorf("loading netrc file: %w", err)
				}
			}
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return log.Close()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}
