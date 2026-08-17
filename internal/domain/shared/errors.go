package shared

// ErrorKind categoriza um AppError para a camada HTTP decidir status/response.
type ErrorKind string

const (
	KindNotFound     ErrorKind = "not_found"
	KindValidation   ErrorKind = "validation"
	KindConflict     ErrorKind = "conflict"
	KindInternal     ErrorKind = "internal"
	KindForbidden    ErrorKind = "forbidden"
	KindUnauthorized ErrorKind = "unauthorized"
	KindUnavailable  ErrorKind = "unavailable"
)

// AppError é o erro tipado que domain/application retornam quando o erro
// é uma condição de negócio conhecida (não uma falha de infraestrutura crua).
type AppError struct {
	Kind    ErrorKind
	Message string
	Details []string
	Err     error
}

// Error implementa a interface error. Se houver causa encapsulada (Err),
// a mensagem inclui o erro original; caso contrário, só a Message de negócio.
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap expõe a causa encapsulada (Err) pra errors.Is/errors.As enxergarem
// através de fmt.Errorf("...: %w", err) no meio do caminho.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewNotFoundError cria um AppError de Kind not_found (mapeado pra 404 no HTTP).
func NewNotFoundError(msg string) *AppError {
	return &AppError{Kind: KindNotFound, Message: msg}
}

// NewValidationError cria um AppError de Kind validation (mapeado pra 400 no HTTP), sem Details.
func NewValidationError(msg string) *AppError {
	return &AppError{Kind: KindValidation, Message: msg}
}

// NewValidationErrorWithDetails cria um AppError de Kind validation com Details
// (mensagens de validação campo a campo).
func NewValidationErrorWithDetails(msg string, details []string) *AppError {
	return &AppError{Kind: KindValidation, Message: msg, Details: details}
}

// NewConflictError cria um AppError de Kind conflict (mapeado pra 409 no HTTP).
func NewConflictError(msg string) *AppError {
	return &AppError{Kind: KindConflict, Message: msg}
}

// NewInternalError cria um AppError de Kind internal (mapeado pra 500 no HTTP)
// encapsulando o erro de infraestrutura original (err) em Err, a mensagem de
// negócio (msg) é a única exposta na resposta, err nunca vaza pro cliente.
func NewInternalError(msg string, err error) *AppError {
	return &AppError{Kind: KindInternal, Message: msg, Err: err}
}

// NewForbiddenError cria um AppError de Kind forbidden (mapeado pra 403 no HTTP).
func NewForbiddenError(msg string) *AppError {
	return &AppError{Kind: KindForbidden, Message: msg}
}

// NewUnauthorizedError cria um AppError de Kind unauthorized (mapeado pra 401 no HTTP).
func NewUnauthorizedError(msg string) *AppError {
	return &AppError{Kind: KindUnauthorized, Message: msg}
}

// NewUnavailableError cria um AppError de Kind unavailable (mapeado pra 503
// no HTTP), dependência externa (banco, fila) fora do ar, não é erro de
// negócio, mas a resposta segue o mesmo formato de ErrorResponse.
func NewUnavailableError(msg string, err error) *AppError {
	return &AppError{Kind: KindUnavailable, Message: msg, Err: err}
}

// NewInternalErrorCustom cria um AppError de Kind internal com mensagem própria,
// usado por erros estáticos definidos como var (ex. sentinel errors de infra).
func NewInternalErrorCustom(msg string) *AppError {
	return &AppError{Kind: KindInternal, Message: msg}
}
