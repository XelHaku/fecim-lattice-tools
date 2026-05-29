package viewmodel

import (
	"fmt"
	"strconv"
	"strings"
)

// PayloadString returns a required string from an action payload map.
func PayloadString(payload map[string]string, key string) (string, error) {
	value, ok := payload[key]
	if !ok {
		return "", fmt.Errorf("missing payload %s", key)
	}
	return value, nil
}

// PayloadStringIn reports whether value matches one of the allowed values.
func PayloadStringIn(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

// PayloadFloat parses a required float64 from an action payload map.
func PayloadFloat(payload map[string]string, key string) (float64, error) {
	value, err := PayloadString(payload, key)
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid payload %s %q: %w", key, value, err)
	}
	return f, nil
}

// OptionalPayloadFloat parses an optional float64 payload value, returning defaultValue when absent.
func OptionalPayloadFloat(payload map[string]string, key string, defaultValue float64) (float64, error) {
	if _, ok := payload[key]; !ok {
		return defaultValue, nil
	}
	return PayloadFloat(payload, key)
}

// PayloadInt parses a required integer from an action payload map.
func PayloadInt(payload map[string]string, key string) (int, error) {
	value, ok := payload[key]
	if !ok {
		return 0, fmt.Errorf("missing payload %s", key)
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid payload %s %q: %w", key, value, err)
	}
	return n, nil
}

// OptionalPayloadInt parses an optional integer payload value, returning defaultValue when absent.
func OptionalPayloadInt(payload map[string]string, key string, defaultValue int) (int, error) {
	if _, ok := payload[key]; !ok {
		return defaultValue, nil
	}
	return PayloadInt(payload, key)
}
