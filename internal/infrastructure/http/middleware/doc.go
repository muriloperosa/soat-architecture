// Package middleware contém os middlewares HTTP usados pelas rotas Gin da API.
//
// authentication_middleware valida o token JWT recebido no cabeçalho das
// requisições e bloqueia o acesso às rotas protegidas quando ele é inválido
// ou está ausente. error_mapper intercepta os erros retornados pelos
// handlers e casos de uso, traduzindo erros de domínio (não encontrado,
// validação, conflito etc) para o status HTTP e corpo de resposta
// correspondentes.
package middleware
