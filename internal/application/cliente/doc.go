// Package cliente reúne os casos de uso da aplicação para o agregado Cliente
// (cadastro, busca e atualização dos dados de quem solicita serviço na
// oficina).
//
// Cada caso de uso orquestra as regras do internal/domain/cliente, validando
// entrada, chamando o repositório e traduzindo o resultado em DTOs de saída,
// sem conter regra de negócio própria. Segue o mesmo padrão estrutural de
// internal/application/ordemservico.
package cliente
