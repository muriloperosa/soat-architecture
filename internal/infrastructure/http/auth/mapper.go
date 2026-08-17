package auth

import appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"

// toLoginInput converte o DTO HTTP de login pro DTO de entrada do LoginUseCase.
func toLoginInput(req LoginRequest) appauth.LoginInput {
	return appauth.LoginInput{Email: req.Email, Senha: req.Senha}
}

// toRefreshInput converte o DTO HTTP de refresh pro DTO de entrada do RefreshUseCase.
func toRefreshInput(req RefreshRequest) appauth.RefreshInput {
	return appauth.RefreshInput{RefreshTokenBruto: req.RefreshToken}
}

// toLogoutInput converte o DTO HTTP de logout (mesmo formato do refresh) pro
// DTO de entrada do LogoutUseCase.
func toLogoutInput(req RefreshRequest) appauth.LogoutInput {
	return appauth.LogoutInput{RefreshTokenBruto: req.RefreshToken}
}

// toTokenResponse monta o DTO HTTP de resposta comum a login e refresh.
func toTokenResponse(accessToken, refreshToken string, expiresIn, refreshTokenExpiresIn int64, requerAlterarSenha bool) TokenResponse {
	return TokenResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresIn:  expiresIn,
		RefreshTokenExpiresIn: refreshTokenExpiresIn,
		RequerAlterarSenha:    requerAlterarSenha,
	}
}
