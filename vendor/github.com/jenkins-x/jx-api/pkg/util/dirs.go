package util

import (
	"os"
	"path/filepath"
)

// JXBinLocation finds the JX config directory and creates a bin directory inside it if it does not already exist. Returns the JX bin path
func JXBinLocation() (string, error) {
	h, err := ConfigDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(h, "bin")
	err = os.MkdirAll(path, DefaultWritePermissions)
	if err != nil {
		return "", err
	}
	return path, nil
}

func ConfigDir() (string, error) {
	path := os.Getenv("JX_HOME")
	if path != "" {
		return path, nil
	}
	h := HomeDir()
	path = filepath.Join(h, ".jx")
	err := os.MkdirAll(path, DefaultWritePermissions)
	if err != nil {
		return "", err
	}
	return path, nil
}

func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	h := os.Getenv("USERPROFILE") // windows
	if h == "" {
		h = "."
	}
	return h
}
