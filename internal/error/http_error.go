package error

// HTTPError represents a custom error type with a status code
type HTTPError struct {
	Message    string
	StatusCode int
}

func (e HTTPError) Error() string {
	return e.Message
}
