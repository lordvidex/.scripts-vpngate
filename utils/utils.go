package utils

import (
	"os"
	"strings"
)

// ExpandHome replaces tilda with UserHomeDir in dir
func ExpandHome(dir string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return strings.ReplaceAll(dir, "~", homeDir)
}

// RemoveExtension removes ovpn, tblk and conf from the filenames and directories
// of the files so that the names are returned without the extensions
func RemoveExtension(filename string) string {
	filename = strings.TrimSpace(filename)
	extensions := []string{".ovpn", ".conf", ".tblk"}
	for _, ext := range extensions {
		filename = strings.TrimSuffix(filename, ext)
	}
	return filename
}
