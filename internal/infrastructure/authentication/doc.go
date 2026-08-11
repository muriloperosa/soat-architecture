// Package authentication contém a geração e validação de tokens JWT usados
// para autenticar as requisições à API, com a biblioteca golang-jwt/v5.
//
// É responsável por emitir o token no login, incluindo as claims necessárias,
// e por validar assinatura e expiração do token recebido nas rotas
// protegidas, sendo consumido pelo middleware de autenticação em
// internal/infrastructure/http/middleware.
package authentication
