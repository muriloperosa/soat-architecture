// Package veiculo reúne os casos de uso da aplicação para o agregado Veículo
// (cadastro e consulta dos veículos vinculados a um Cliente, que dão origem
// a uma Ordem de Serviço).
//
// Cada caso de uso orquestra as regras do internal/domain/veiculo, validando
// entrada, chamando o repositório e traduzindo o resultado em DTOs de saída,
// sem conter regra de negócio própria. Segue o mesmo padrão estrutural de
// internal/application/ordemservico.
package veiculo
