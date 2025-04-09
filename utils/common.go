package utils

import "regexp"

func IsAllDigits(s string) bool {
	pattern := regexp.MustCompile(`^[0-9]+$`)
	return pattern.MatchString(s)
}
