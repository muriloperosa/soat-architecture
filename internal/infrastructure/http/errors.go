package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

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

	c.JSON(status, gin.H{
		"type":    string(appErr.Kind),
		"message": message,
		"details": appErr.Details,
	})
}
