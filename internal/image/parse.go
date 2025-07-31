package image

import (
	"errors"
	"strings"

	"github.com/jcchavezs/nuro/internal/log"
	"go.uber.org/zap"
)

const DockerRegistry = "registry-1.docker.io"

// ParseImage parses an image name
func ParseImage(image string) (registry string, name string, tag string, digest string, err error) {
	if strings.HasPrefix(image, ":") || strings.HasPrefix(image, "@") || strings.HasPrefix(image, "/") {
		err = errors.New("invalid image")
		return
	}

	fields := []zap.Field{}

	switch strings.Count(image, "/") {
	default:
		err = errors.New("invalid image")
		return
	case 0:
		registry = DockerRegistry
		image = "library/" + image
	case 1:
		registry = DockerRegistry
	case 2:
		registry, image, _ = strings.Cut(image, "/")
	}

	if registry == "docker.io" {
		registry = DockerRegistry
	}

	fields = append(fields, zap.String("registry", registry))

	if strings.Contains(image, "@") {
		fields = append(fields, zap.String("digest", digest))
		image, digest, _ = strings.Cut(image, "@")
	}

	if strings.Contains(image, ":") {
		image, tag, _ = strings.Cut(image, ":")
	}

	if digest == "" && tag == "" {
		tag = "latest"
	}

	name = image
	fields = append(fields, zap.String("name", name), zap.String("tag", tag))
	log.Logger.Debug("Parsing image", fields...)

	return
}
