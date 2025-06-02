package utils

import (
	"hstin/geocoder/mapping"
	"strconv"
	"strings"
)

func ParseStringAsNumber(s string) int64 {
	// Directly handle potential mappings
	for word, num := range mapping.NumberWords {
		if strings.Contains(strings.ToLower(s), word) {
			return int64(num)
		}
	}

	// Clean the string by removing non-numeric characters
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)

	if cleaned == "" {
		return 0
	}

	number, err := strconv.Atoi(cleaned)
	if err != nil {
		return 0
	}

	return int64(number)
}
