package utils

import (
	"fmt"
	"strings"
)

const (
	NoRouteToHostErr    = "no route to host"
	NotFoundErr         = "not found"
	UnauthorizedErr     = "Unauthorized"
	ForbiddenErr        = "forbidden"
	ConnexionRefusedErr = "connection refused"
)

func IsUnauthorizedError(message string, object string) (bool, error) {
	//TODO: update the error message
	if strings.Contains(message, UnauthorizedErr) {
		return true, fmt.Errorf("cannot access %s", object)
	}
	return false, nil
}

func IsForbiddenError(message string, object string) (bool, error) {
	if strings.Contains(message, ForbiddenErr) {
		return true, fmt.Errorf("cannot access %s", object)
	}
	return false, nil
}

func IsNotFoundError(message string) bool {
	return strings.Contains(message, NotFoundErr)
}

func IsNoRouteToHostError(message string) bool {
	return strings.Contains(message, NoRouteToHostErr)
}

func IsConnexionRefusedError(message string) bool {
	return strings.Contains(message, ConnexionRefusedErr)
}
