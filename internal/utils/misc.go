package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/Data-Corruption/blog"
)

// Contains checks if a slice contains a given element.
func Contains[T comparable](element T, slice []T) bool {
	for _, sliceElement := range slice {
		if element == sliceElement {
			return true
		}
	}
	return false
}

// Behold... my unholy attack on god himself.
func Ternary[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

// PromptString prompts the user for a string. Is blocking.
func PromptInt(prompt string) int {
	var i int
	fmt.Print(prompt)
	fmt.Scan(&i)
	return i
}

// ArgPresent checks if a given argument is present in the command line arguments.
func ArgPresent(arg string) bool {
	for _, a := range os.Args {
		if a == arg {
			return true
		}
	}
	return false
}

// GenRandomString generates a cryptographically secure random token of the given size.
// Output is URL and filename safe.
func GenRandomString(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetPublicIP gets the public IP of the machine using https://api.ipify.org
func GetPublicIP() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return "", fmt.Errorf("error fetching IP address: %w", err)
	}
	defer resp.Body.Close()
	// handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response: %d %s", resp.StatusCode, resp.Status)
	}
	// read then return the response
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}
	return string(ip), nil
}

func TailwindInstalled() bool {
	cmd := exec.Command("npx", "tailwindcss", "--help")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		blog.Errorf("TailwindCSS is not installed: %v", err)
		return false
	} else {
		blog.Info("TailwindCSS is installed")
		return true
	}
}
