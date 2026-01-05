package error

// CustomError represents a custom error type with a status code
type CustomError struct {
	Message    string
	StatusCode int
}

func (e CustomError) Error() string {
	return e.Message
}
