package relatorio

type ConsultarTransicaoStatusRequest struct {
	StartDate  string `form:"start_date" binding:"required"`
	FinalDate  string `form:"final_date" binding:"required"`
	FromStatus string `form:"from_status" binding:"required"`
	ToStatus   string `form:"to_status" binding:"required"`
	Unit       string `form:"unit"`
}

type TransicaoStatusResponse struct {
	TotalOrdensServico int     `json:"total_ordens_servico" example:"12"`
	TempoMedio         float64 `json:"tempo_medio" example:"1.5"`
	TempoMinimo        float64 `json:"tempo_minimo" example:"0.5"`
	TempoMaximo        float64 `json:"tempo_maximo" example:"3"`
	Unidade            string  `json:"unidade" example:"h"`
}
