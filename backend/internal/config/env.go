package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var envKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// LoadLocalEnv reads the repository-root .env for local development.
// Environment variables already supplied by the process take precedence.
// The file is optional: deployment environments should inject variables directly.
func LoadLocalEnv() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	for _, candidate := range []string{dir, filepath.Dir(dir)} {
		// Find the repository root; do not accidentally load backend/.env.
		if _, err := os.Stat(filepath.Join(candidate, "compose.yaml")); err != nil {
			continue
		}
		path := filepath.Join(candidate, ".env")
		file, err := os.Open(path)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("open local .env: %w", err)
		}
		defer file.Close()
		return LoadReader(file, os.LookupEnv, os.Setenv)
	}
	return nil
}

// LoadReader parses a small, predictable KEY=VALUE subset of dotenv syntax.
// Blank lines and full-line comments are ignored. Single/double quoted values
// and values containing '=' are supported; shell expansion is intentionally not.
// Neither errors nor logs contain secret values.
func LoadReader(reader io.Reader, lookup func(string) (string, bool), setenv func(string, string) error) error {
	scanner := bufio.NewScanner(reader)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || !envKey.MatchString(key) {
			return fmt.Errorf("invalid .env entry at line %d", lineNumber)
		}
		value = strings.TrimSpace(value)
		if strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
			quote := value[:1]
			if len(value) < 2 || !strings.HasSuffix(value, quote) {
				return fmt.Errorf("invalid quoted .env entry at line %d", lineNumber)
			}
			value = value[1 : len(value)-1]
		}
		if _, exists := lookup(key); exists {
			continue
		}
		if err := setenv(key, value); err != nil {
			return fmt.Errorf("set environment variable at line %d: %w", lineNumber, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read .env: %w", err)
	}
	return nil
}

// Port returns a validated HTTP listen port.
func Port(getenv func(string) string) (string, error) {
	value := getenv("PORT")
	if value == "" {
		return "8080", nil
	}
	number, err := strconv.Atoi(value)
	if err != nil || number < 1 || number > 65535 {
		return "", errors.New("PORT must be between 1 and 65535")
	}
	return value, nil
}
