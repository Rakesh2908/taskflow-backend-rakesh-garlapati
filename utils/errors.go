package utils

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDB                 = errors.New("database error")
)

// DBError intentionally hides the underlying driver error from callers.
// It unwraps only to ErrDB, while repositories should log the original error.
type DBError struct {
	Op string
}

func (e *DBError) Error() string {
	if e.Op == "" {
		return ErrDB.Error()
	}
	return "db: " + e.Op
}

func (e *DBError) Unwrap() error { return ErrDB }

func WrapDB(op string) error {
	return &DBError{Op: op}
}

