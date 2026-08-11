// Package peca reúne os casos de uso da aplicação para o agregado Peça
// (o estoque de peças que podem ser aplicadas numa Ordem de Serviço, com
// preço e quantidade disponível).
//
// Cada caso de uso orquestra as regras do internal/domain/peca, validando
// entrada, chamando o repositório e traduzindo o resultado em DTOs de saída,
// sem conter regra de negócio própria. Segue o mesmo padrão estrutural de
// internal/application/ordemservico.
package peca
