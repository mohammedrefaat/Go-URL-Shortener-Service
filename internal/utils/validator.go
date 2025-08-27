package utils

import (
	"net/url"
	"strings"
)

var maliciousDomains = []string{
	"malware.example.com",
	"phishing.example.com",
	// Add more malicious domains
}

func IsValidURL(rawURL string) bool {
	// Parse URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Check scheme
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	// Check if host exists
	if u.Host == "" {
		return false
	}

	// Check against malicious domains
	for _, domain := range maliciousDomains {
		if strings.Contains(u.Host, domain) {
			return false
		}
	}

	return true
}
