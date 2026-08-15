package auth

// LoginInput é o DTO de entrada do LoginUseCase.
type LoginInput struct {
	Email string
	Senha string
}

// LoginOutput é o DTO de saída do LoginUseCase.
type LoginOutput struct {
	AccessToken  string
	RefreshToken string
}

// RefreshInput é o DTO de entrada do RefreshUseCase.
type RefreshInput struct {
	RefreshTokenBruto string
}

// RefreshOutput é o DTO de saída do RefreshUseCase.
type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
}

// LogoutInput é o DTO de entrada do LogoutUseCase.
type LogoutInput struct {
	RefreshTokenBruto string
}
