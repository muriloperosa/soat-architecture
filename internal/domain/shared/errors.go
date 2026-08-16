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

// NewNotFoundError cria um AppError do tipo not_found, com mensagem adicional.
func NewNotFoundError(msg string) *AppError {
	return &AppError{Kind: KindNotFound, Message: msg}
}

// NewValidationError cria um AppError do tipo validation, com mensagem adicional.
func NewValidationError(msg string) *AppError {
	return &AppError{Kind: KindValidation, Message: msg}
}

// NewValidationErrorWithDetails cria um AppError do tipo validation, com mensagem adicional e detalhes.
func NewValidationErrorWithDetails(msg string, details []string) *AppError {
	return &AppError{Kind: KindValidation, Message: msg, Details: details}
}

// NewConflictError cria um AppError do tipo conflict, com mensagem adicional.
func NewConflictError(msg string) *AppError {
	return &AppError{Kind: KindConflict, Message: msg}
}

// NewInternalError cria um AppError do tipo internal, com mensagem adicional e erro interno.
func NewInternalError(msg string, err error) *AppError {
	return &AppError{Kind: KindInternal, Message: msg, Err: err}
}

// NewInternalErrorCustom cria um AppError do tipo internal sem mensagem adicional, apenas com a Kind.
func NewInternalErrorCustom(msg string) *AppError {
	return &AppError{Kind: KindInternal}
}
