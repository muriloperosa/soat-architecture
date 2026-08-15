package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// ErrorResponse é o corpo de erro emitido por RespondError e pelas validações
// de binding dos handlers (Details vazio nesse segundo caso).
type ErrorResponse struct {
	// Type é o shared.ErrorKind do erro (validation, unauthorized, forbidden, not_found, conflict, internal)
	Type string `json:"type" example:"validation"`
	// Message é a mensagem legível pro cliente
	Message string `json:"message" example:"corpo inválido"`
	// Details lista mensagens de validação campo a campo, quando aplicável
	Details []string `json:"details,omitempty"`
}

var statusByKind = map[shared.ErrorKind]int{
	shared.KindNotFound:     http.StatusNotFound,
	shared.KindValidation:   http.StatusBadRequest,
	shared.KindConflict:     http.StatusConflict,
	shared.KindInternal:     http.StatusInternalServerError,
	shared.KindForbidden:    http.StatusForbidden,
	shared.KindUnauthorized: http.StatusUnauthorized,
}

// RespondError traduz um erro de domínio/aplicação para resposta HTTP.
// Erros que não são *shared.AppError são tratados como internos, sem
// vazar detalhes de infraestrutura na resposta.
func RespondError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if !errors.As(err, &appErr) {
		appErr = shared.NewInternalError("erro interno", err)
	}

	status, ok := statusByKind[appErr.Kind]
	if !ok {
		status = http.StatusInternalServerError
	}

	message := appErr.Message
	if appErr.Kind == shared.KindInternal {
		message = "erro interno"
	}

	c.JSON(status, ErrorResponse{
		Type:    string(appErr.Kind),
		Message: message,
		Details: appErr.Details,
	})
}

// RespondValidationError responde 400 (ErrorKind validation) — atalho pra
// erros de binding/validação de corpo HTTP, antes de chegar num use case.
func RespondValidationError(c *gin.Context, message string) {
	RespondError(c, shared.NewValidationError(message))
}

// RespondNotFoundError responde 404 (ErrorKind not_found).
func RespondNotFoundError(c *gin.Context, message string) {
	RespondError(c, shared.NewNotFoundError(message))
}

// RespondConflictError responde 409 (ErrorKind conflict).
func RespondConflictError(c *gin.Context, message string) {
	RespondError(c, shared.NewConflictError(message))
}

// RespondForbiddenError responde 403 (ErrorKind forbidden).
func RespondForbiddenError(c *gin.Context, message string) {
	RespondError(c, shared.NewForbiddenError(message))
}

// RespondUnauthorizedError responde 401 (ErrorKind unauthorized).
func RespondUnauthorizedError(c *gin.Context, message string) {
	RespondError(c, shared.NewUnauthorizedError(message))
}

// RespondInternalError responde 500 (ErrorKind internal), sem vazar o erro
// original (err) na resposta — só a mensagem genérica.
func RespondInternalError(c *gin.Context, message string, err error) {
	RespondError(c, shared.NewInternalError(message, err))
}
