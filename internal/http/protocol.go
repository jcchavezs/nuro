package http

func ResolveProtocol(insecure bool) string {
	if insecure {
		return "http"
	}

	return "https"
}
