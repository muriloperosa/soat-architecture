// Package servico contém a implementação de persistência do agregado
// Serviço em MySQL, via GORM.
//
// O model representa a tabela no banco, o mapper converte entre o model e
// a entidade de internal/domain/servico, e o repository implementa a
// interface de persistência definida no domínio usando esse model e mapper.
package servico
