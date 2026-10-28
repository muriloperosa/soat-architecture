package auth

import (
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// AppClaims é a identidade carregada por um access token já validado: quem é
// (Subject, o ID do usuário), de que fonte de identidade (Tipo) e com qual
// papel. Jti identifica esse access token especificamente (não confundir com
// o refresh token bruto) e é usado na revogação.
//
// O formato do token (JWT, claims registrados, expiração) é detalhe do
// adaptador em internal/infrastructure/auth, que converte para este tipo.
type AppClaims struct {
	Subject string
	Tipo    TipoUsuario
	Papel   shared.PapelUsuario
	Jti     string
}
