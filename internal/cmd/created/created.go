package created

import (
	"errors"
	"fmt"
	"time"

	"github.com/jcchavezs/nuro/internal/api/blob"
	"github.com/jcchavezs/nuro/internal/api/manifest"
	"github.com/jcchavezs/nuro/internal/auth"
	"github.com/jcchavezs/nuro/internal/image"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.PersistentFlags().Bool("insecure", false, "Allow communication with an insecure registry")
}

var RootCmd = &cobra.Command{
	Use:     "created <image>",
	Short:   "Shows the creation date for a given image",
	Example: "$ nuro created alpine",
	Args:    cobra.ExactArgs(1),
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

		created, ok := resolveDateFromConfig(cfg)
		if !ok {
			return errors.New("no creation date found")
		}

		if _, err := fmt.Fprint(cmd.OutOrStdout(), created); err != nil {
			return fmt.Errorf("writing to stdout: %w", err)
		}

		return nil
	},
}

func resolveDateFromConfig(cfg *blob.ConfigBlob) (string, bool) {
	var created string
	if cfg.Created.IsZero() {
		var l map[string]string
		if len(cfg.Annotations) != 0 {
			l = cfg.Annotations
		} else if len(cfg.Config.Labels) != 0 {
			l = cfg.Config.Labels
		} else {
			return "", false
		}

		created = l["org.opencontainers.image.created"]
	} else {
		created = cfg.Created.Format(time.RFC3339)
	}

	return created, true
}
