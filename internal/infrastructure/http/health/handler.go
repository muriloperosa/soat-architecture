package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
)

const msgBancoIndisponivel = "Banco de dados relacional indisponível."

// NewHealthHandler retorna um handler Gin que faz ping no banco via GORM.
//
// @Summary Verifica saúde da API
// @Description Retorna 200 se a API e a conexão com o MySQL estão respondendo
// @Tags Health
// @Produce json
// @Success 200 {object} HealthCheckResponse
// @Failure 503 {object} httperror.ErrorResponse
// @Router /v1/health [get]
func NewHealthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			httperror.RespondUnavailableError(c, msgBancoIndisponivel, err)
			return
		}

		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			httperror.RespondUnavailableError(c, msgBancoIndisponivel, err)
			return
		}

		c.JSON(http.StatusOK, HealthCheckResponse{Status: "ok"})
	}
}
