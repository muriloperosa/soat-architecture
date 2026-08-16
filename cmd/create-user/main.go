package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	appusuario "github.com/muriloperosa/soat-architecture/internal/application/usuario"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// create-user cria um usuário interno direto no banco, sem passar pelo
// HTTP/JWT. Resolve o bootstrap do primeiro admin (que precisaria de um
// admin ja existente pra bater em POST /v1/usuarios) e serve pra testar o
// dominio manualmente.
func main() {
	nome := flag.String("nome", "", "nome do usuario")
	email := flag.String("email", "", "email do usuario")
	senha := flag.String("senha", "", "senha inicial (provisoria)")
	papel := flag.String("papel", "ADMINISTRADOR", "papel: ADMINISTRADOR, MECANICO ou ATENDENTE")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("erro ao carregar config: %v", err)
	}

	db, err := mysql.NewMigrationConnection(cfg)
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}

	c := wiring.NewContainer(cfg, db)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := c.CriarUsuarioUC.Executar(ctx, appusuario.CriarUsuarioInput{
		Nome:         *nome,
		Email:        *email,
		SenhaInicial: *senha,
		Papel:        shared.PapelUsuario(*papel),
	})
	if err != nil {
		log.Fatalf("erro ao criar usuario: %v", err)
	}

	fmt.Printf("usuario criado: id=%d email=%s papel=%s\n", out.ID, out.Email, out.Papel)
}
