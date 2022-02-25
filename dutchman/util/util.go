package util

import (
	"errors"
	"strings"
)

func ExtractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("token not found")
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return "", errors.New("invalid token")
	}

	if headerParts[0] != "Bearer" {
		return "", errors.New("invalid token")
	}

	return headerParts[1], nil
}
