// Package veiculo contém a implementação de persistência do agregado
// Veículo em MySQL, via GORM.
//
// O model representa a tabela no banco, o mapper converte entre o model e
// a entidade de internal/domain/veiculo, e o repository implementa a
// interface de persistência definida no domínio usando esse model e mapper.
package veiculo
