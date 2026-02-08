package errors

// CodedError represents an error with a code and message.
// This consolidates duplicate error types across packages (AnalysisError, DeckError, EventError).
type CodedError struct {
	Code    string
	Message string
}

func (e *CodedError) Error() string {
	return e.Message
}

// New creates a new CodedError with the given code and message.
func New(code, message string) *CodedError {
	return &CodedError{
		Code:    code,
		Message: message,
	}
}
