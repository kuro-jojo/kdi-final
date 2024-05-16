package utils

import (
	"fmt"
	"strings"
)

const (
	// ErrDuplicateKey is the error message for duplicate key
	ErrDuplicateKey = "duplicate key"

	// ErrNotFound is the error message for not found
	ErrNotFound = "not found"

	// ErrSameValue is the error message for same value
	ErrSameValue = "same value"
)

func OnDuplicateKeyError(dbErr error, entity string) error {
	if strings.Contains(dbErr.Error(), ErrDuplicateKey) {
		return fmt.Errorf("%s already exists", entity)
	}
	return nil
}
func OnNotFoundError(dbErr error, entity string) error {
	if strings.Contains(dbErr.Error(), ErrNotFound) {
		return fmt.Errorf("%s not found", entity)
	}
	return nil
}

func OnSameValueError(dbErr error, entity string) error {
	if strings.Contains(dbErr.Error(), ErrSameValue) {
		return fmt.Errorf("use a different value for the %s", entity)
	}
	return nil
}

// func HandleCreationError(e error) error {
// 	if e == nil {
// 		return nil
// 	}
// 	err := OnDuplicateKey(e, "project")
// 	// if err != nil {
// 	// 	// other error
// 	// }
// 	return err
// }
