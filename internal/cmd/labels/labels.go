package labels

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jcchavezs/nuro/internal/api/blob"
	"github.com/jcchavezs/nuro/internal/api/manifest"
	"github.com/jcchavezs/nuro/internal/auth"
	"github.com/jcchavezs/nuro/internal/image"
	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
)

var Formats = map[OutputFormat][]string{
	Table: {"table"},
	JSON:  {"json"},
}

type OutputFormat int

const (
	Table OutputFormat = iota
	JSON
)

var outputFormat OutputFormat = Table

func init() {
	RootCmd.PersistentFlags().Bool("insecure", false, "Allow communication with an insecure registry")
	RootCmd.Flags().Var(
		enumflag.New(&outputFormat, "string", Formats, enumflag.EnumCaseInsensitive),
		"output",
		"Sets the output format",
	)
}

var RootCmd = &cobra.Command{
	Use:     "labels <image>",
	Short:   "Shows labels for a given image",
	Args:    cobra.ExactArgs(1),
	Example: "$ nuro labels alpine:3.18.12",
	RunE: func(cmd *cobra.Command, args []string) error {
		registry, name, tag, digest, err := image.ParseImage(args[0])
		if err != nil {
			return fmt.Errorf("parsing image: %w", err)
		}

		reference := digest
		if reference == "" {
			reference = tag
		}

		ctx := auth.InjectImageMetadata(cmd.Context(), auth.ImageMetadata{Registry: registry, Name: name})

		insecure, err := cmd.Flags().GetBool("insecure")
		if err != nil {
			return fmt.Errorf("getting insecure flag: %w", err)
		}

		d, err := manifest.GetConfigDigestFromManifest(ctx, registry, insecure, name, reference)
		if err != nil {
			return fmt.Errorf("getting config digest from manifest: %w", err)
		}

		cfg, err := blob.GetConfigBlob(ctx, registry, insecure, name, d)
		if err != nil {
			return fmt.Errorf("getting labels from config blob: %w", err)
		}

		var l map[string]string
		if len(cfg.Annotations) != 0 {
			l = cfg.Annotations
		} else if len(cfg.Config.Labels) != 0 {
			l = cfg.Config.Labels
		} else {
			return errors.New("no labels found")
		}

		switch outputFormat {
		case JSON:
			if err = json.NewEncoder(cmd.OutOrStdout()).Encode(l); err != nil {
				return fmt.Errorf("writing to stdout: %w", err)
			}
		default:
			t := table.NewWriter()
			t.SetOutputMirror(cmd.OutOrStdout())
			t.AppendHeader(table.Row{"Key", "Value"})
			for k, v := range l {
				t.AppendRow(table.Row{k, text.WrapSoft(v, 60)})
			}
			t.Render()
		}

		return nil
	},
}
