// Package utils holds common data structures and functions useful for working
// with OpenChirp.
package utils

import (
	"encoding/csv"
	"strconv"
	"strings"
)

// ParseCSVConfig parses a single config field that follows comma and
// optional quotes seperated syntax into it's constituent tokens
// Possible errors returned are from the encoding/csv package.
// The error can be referenced by it's concrete type *csv.ParseError,
// which can give useful information about where the parse error occurred.
// Example: errColumn := err.(*csv.ParseError).Column
func ParseCSVConfig(configline string) ([]string, error) {
	r := csv.NewReader(strings.NewReader(configline))
	r.TrimLeadingSpace = true
	// Call Read only once because there should only be one line
	tokens, err := r.Read()
	return tokens, err
}

// ParseOCValue tries to parse the three typical primitive data types used
// in OpenChirp. It first tries to parse the value as a float64. Then,
// it tries to parse as a bool. If all else fails, it returns the value
// as a string.
func ParseOCValue(value string) interface{} {
	// Try float64
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return v
	}

	// Try bool
	if v, err := strconv.ParseBool(value); err == nil {
		return v
	}

	// Take as string
	return value
}
