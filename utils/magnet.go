package utils

import (
	"errors"
	"regexp"
)

func ExtractMagnetHash(magnet string) (string, error) {
	hashRegex := regexp.MustCompile("btih:([a-fA-F0-9]+)")
	match := hashRegex.FindStringSubmatch(magnet)
	if len(match) != 2 {
		return "", errors.New("invalid magnet link")
	}
	return match[1], nil
}
