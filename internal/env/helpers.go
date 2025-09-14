package env

import (
	"os"
	"strings"
)

// Remove surrounding quotes
// Check if string starts and ends with either single or double quotes.
// if it is, then remove it. Else return original string
func removeStringQuotes(value string) string {
	if len(value) > 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}

	return value
}

// Process env file lines
// Trim whitespaces and check if process is a valid key-value pair or
// a comment. If it is a valid pair, Set it into os environment variables
// Ignore malformed lines.
func processEnvFileLine(line string) error {
	line = strings.TrimSpace(line)

	if (line == "") || strings.HasPrefix(line, "#") {
		return nil
	}

	if !(strings.HasPrefix(line, "export ")) {
		return nil
	}

	line = strings.TrimPrefix(line, "export ")
	parts := strings.SplitN(line, "=", 2)

	if len(parts) != 2 {
		return nil
	}

	key := strings.TrimSpace(parts[0])
	value := removeStringQuotes(strings.TrimSpace(parts[1]))

	err := os.Setenv(key, value)
	if err != nil {
		return err
	}

	return nil
}
