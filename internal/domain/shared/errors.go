package shared

// ErrorKind categoriza um AppError para a camada HTTP decidir status/response.
type ErrorKind string

const (
	KindNotFound   ErrorKind = "not_found"
	KindValidation ErrorKind = "validation"
	KindConflict   ErrorKind = "conflict"
	KindInternal   ErrorKind = "internal"
)

// AppError é o erro tipado que domain/application retornam quando o erro
// é uma condição de negócio conhecida (não uma falha de infraestrutura crua).
type AppError struct {
	Kind    ErrorKind
	Message string
	Details []string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(msg string) *AppError {
	return &AppError{Kind: KindNotFound, Message: msg}
}

func NewValidationError(msg string) *AppError {
	return &AppError{Kind: KindValidation, Message: msg}
}

func NewValidationErrorWithDetails(msg string, details []string) *AppError {
	return &AppError{Kind: KindValidation, Message: msg, Details: details}
}

func NewConflictError(msg string) *AppError {
	return &AppError{Kind: KindConflict, Message: msg}
}

func NewInternalError(msg string, err error) *AppError {
	return &AppError{Kind: KindInternal, Message: msg, Err: err}
}
