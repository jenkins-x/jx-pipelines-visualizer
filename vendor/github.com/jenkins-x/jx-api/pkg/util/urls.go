package util

import (
	"net/url"
	"strings"
)

// SanitizeURL sanitizes by stripping the user and password
func SanitizeURL(unsanitizedURL string) string {
	u, err := url.Parse(unsanitizedURL)
	if err != nil {
		return unsanitizedURL
	}
	return stripCredentialsFromURL(u)
}

// stripCredentialsFromURL strip credentials from URL
func stripCredentialsFromURL(u *url.URL) string {
	pass, hasPassword := u.User.Password()
	userName := u.User.Username()
	if hasPassword {
		textToReplace := pass + "@"
		textToReplace = ":" + textToReplace
		if userName != "" {
			textToReplace = userName + textToReplace
		}
		return strings.Replace(u.String(), textToReplace, "", 1)
	}
	return u.String()
}
