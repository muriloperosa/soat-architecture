// Package servico reúne os casos de uso da aplicação para o agregado Serviço
// (o catálogo de serviços que podem ser oferecidos numa Ordem de Serviço,
// como troca de óleo, alinhamento, revisão etc).
//
// Cada caso de uso orquestra as regras do internal/domain/servico, validando
// entrada, chamando o repositório e traduzindo o resultado em DTOs de saída,
// sem conter regra de negócio própria. Segue o mesmo padrão estrutural de
// internal/application/ordemservico.
package servico
