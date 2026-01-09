package error

// HttpError represents a custom error type with a status code
type HttpError struct {
	Message    string
	StatusCode int
}

func (e HttpError) Error() string {
	return e.Message
}
