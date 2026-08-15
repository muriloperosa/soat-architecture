package health

// HealthCheckResponse é o corpo de sucesso de GET /v1/health.
type HealthCheckResponse struct {
	Status string `json:"status" example:"ok"`
}
