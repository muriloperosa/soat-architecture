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

// NewReadinessHandler é a readiness probe do Kubernetes: o pod só recebe
// tráfego do Service enquanto o banco responde. Mesma verificação de
// NewHealthHandler, exposta numa rota própria para a probe.
//
// @Summary Readiness probe
// @Description Retorna 200 quando a API consegue falar com o MySQL; 503 tira o pod do balanceamento sem reiniciá-lo
// @Tags Health
// @Produce json
// @Success 200 {object} HealthCheckResponse
// @Failure 503 {object} httperror.ErrorResponse
// @Router /v1/health/ready [get]
func NewReadinessHandler(db *gorm.DB) gin.HandlerFunc {
	return NewHealthHandler(db)
}

// NewLivenessHandler é a liveness probe do Kubernetes. Responde 200 enquanto
// o processo atende HTTP e não consulta o banco de propósito: se dependesse
// do MySQL, uma queda do banco faria o kubelet reiniciar todas as réplicas
// em loop sem resolver nada.
//
// @Summary Liveness probe
// @Description Retorna 200 enquanto o processo da API está de pé; não verifica dependências externas
// @Tags Health
// @Produce json
// @Success 200 {object} HealthCheckResponse
// @Router /v1/health/live [get]
func NewLivenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthCheckResponse{Status: "ok"})
	}
}
