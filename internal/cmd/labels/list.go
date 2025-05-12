package labels

import (
	"fmt"

	"github.com/jcchavezs/nuro/internal/auth"
	"github.com/jcchavezs/nuro/internal/image"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List labels for a given image",
	Args:  cobra.ExactArgs(1),
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

		d, err := getConfigDigestFromManifest(ctx, registry, insecure, name, reference)
		if err != nil {
			return fmt.Errorf("getting config digest from manifest: %w", err)
		}

		l, err := getLabelsFromBlob(ctx, registry, insecure, name, d)
		if err != nil {
			return fmt.Errorf("getting labels from config blob: %w", err)
		}

		_, err = cmd.OutOrStdout().Write(l)
		if err != nil {
			return fmt.Errorf("writing to stdout: %w", err)
		}

		return nil
	},
}
