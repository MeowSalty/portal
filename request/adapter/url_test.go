package adapter

import "testing"

func TestJoinBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		endpoint string
		expected string
	}{
		{
			name:     "base without trailing slash, endpoint without leading slash",
			baseURL:  "https://test.com",
			endpoint: "prefix/v1/messages",
			expected: "https://test.com/prefix/v1/messages",
		},
		{
			name:     "base with trailing slash, endpoint without leading slash",
			baseURL:  "https://test.com/",
			endpoint: "prefix/v1/messages",
			expected: "https://test.com/prefix/v1/messages",
		},
		{
			name:     "base without trailing slash, endpoint with leading slash",
			baseURL:  "https://test.com",
			endpoint: "/prefix/v1/messages",
			expected: "https://test.com/prefix/v1/messages",
		},
		{
			name:     "base with trailing slash, endpoint with leading slash",
			baseURL:  "https://test.com/",
			endpoint: "/prefix/v1/messages",
			expected: "https://test.com/prefix/v1/messages",
		},
		{
			name:     "endpoint with double leading slashes",
			baseURL:  "https://test.com/",
			endpoint: "//v1/messages",
			expected: "https://test.com/v1/messages",
		},
		{
			name:     "empty endpoint returns normalized base",
			baseURL:  "https://test.com/",
			endpoint: "",
			expected: "https://test.com",
		},
		{
			name:     "empty baseURL returns normalized endpoint",
			baseURL:  "",
			endpoint: "/v1/messages",
			expected: "/v1/messages",
		},
		{
			name:     "absolute endpoint URL returns endpoint",
			baseURL:  "https://test.com/",
			endpoint: "https://api.example.com/v1/messages",
			expected: "https://api.example.com/v1/messages",
		},
		{
			name:     "prefix endpoint with trailing slash and default endpoint leading slash",
			baseURL:  "https://test.com/",
			endpoint: "prefix//v1/messages",
			expected: "https://test.com/prefix/v1/messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinBaseURL(tt.baseURL, tt.endpoint)
			if result != tt.expected {
				t.Fatalf("joinBaseURL(%q, %q) = %q; want %q", tt.baseURL, tt.endpoint, result, tt.expected)
			}
		})
	}
}
