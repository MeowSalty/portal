package adapter

import "strings"

func joinBaseURL(baseURL, endpoint string) string {
	if endpoint == "" {
		return strings.TrimRight(baseURL, "/")
	}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}

	base := strings.TrimRight(baseURL, "/")
	normalizedEndpoint := strings.TrimLeft(endpoint, "/")
	for strings.Contains(normalizedEndpoint, "//") {
		normalizedEndpoint = strings.ReplaceAll(normalizedEndpoint, "//", "/")
	}

	if base == "" {
		return "/" + normalizedEndpoint
	}

	return base + "/" + normalizedEndpoint
}
