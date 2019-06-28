package libgin

// Common utilities for the GIN services

import (
	"os"
)

// ReadConfDefault returns the value of a configuration env variable.
// If the variable is not set, the default is returned.
func ReadConfDefault(key, defval string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defval
	}
	return value
}

// ReadConf returns the value of a configuration env variable.
// If the variable is not set, an empty string is returned (ignores any errors).
func ReadConf(key string) string {
	value, _ := os.LookupEnv(key)
	return value
}
