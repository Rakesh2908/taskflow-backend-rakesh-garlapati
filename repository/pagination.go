package repository

// Page represents limit/offset pagination.
// If nil is passed to repository methods, they return the full result set.
type Page struct {
	Limit  int
	Offset int
}

