package image

import (
	"encoding/json"
	"fmt"

	"github.com/jcchavezs/nuro/internal/api/blob"
	"github.com/jcchavezs/nuro/internal/api/manifest"
	"github.com/jcchavezs/nuro/internal/auth"
	"github.com/jcchavezs/nuro/internal/image"
	"github.com/spf13/cobra"
)

var labelsCmd = &cobra.Command{
	Use:     "labels",
	Short:   "List labels for a given image",
	Args:    cobra.ExactArgs(1),
	Example: "$ nuro image labels alpine",
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

		l := cfg.Annotations
		if len(l) == 0 {
			l = cfg.Config.Labels
		}

		if err = json.NewEncoder(cmd.OutOrStdout()).Encode(l); err != nil {
			return fmt.Errorf("writing to stdout: %w", err)
		}

		return nil
	},
}
