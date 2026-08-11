// Package ordemservico contém a implementação de persistência do agregado
// Ordem de Serviço em MySQL, via GORM.
//
// O model representa a tabela no banco, incluindo o relacionamento com
// Cliente, Veículo, Serviços e Peças aplicados, o mapper converte entre o
// model e a entidade de internal/domain/ordemservico, e o repository
// implementa a interface de persistência definida no domínio usando esse
// model e mapper.
package ordemservico
