package image

import (
	"errors"
	"strings"
)

const DockerRegistry = "registry-1.docker.io"

func ParseImage(image string) (registry string, name string, tag string, digest string, err error) {
	if strings.HasPrefix(image, ":") || strings.HasPrefix(image, "@") || strings.HasPrefix(image, "/") {
		err = errors.New("invalid image")
		return
	}

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

	if strings.Contains(image, "@") {
		image, digest, _ = strings.Cut(image, "@")
	}

	if strings.Contains(image, ":") {
		image, tag, _ = strings.Cut(image, ":")
	}

	if digest == "" && tag == "" {
		tag = "latest"
	}

	name = image

	return
}
