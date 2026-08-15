package http

import appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"

func toLoginInput(req LoginRequest) appauth.LoginInput {
	return appauth.LoginInput{Email: req.Email, Senha: req.Senha}
}

func toRefreshInput(req RefreshRequest) appauth.RefreshInput {
	return appauth.RefreshInput{RefreshTokenBruto: req.RefreshToken}
}

func toLogoutInput(req RefreshRequest) appauth.LogoutInput {
	return appauth.LogoutInput{RefreshTokenBruto: req.RefreshToken}
}

func toTokenResponse(accessToken, refreshToken string) TokenResponse {
	return TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}
}
