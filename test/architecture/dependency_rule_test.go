// Package architecture_test verifica a regra de dependência da Clean
// Architecture (docs/adr/0007-adocao-clean-architecture.md) lendo os imports dos
// arquivos Go. Roda junto com `go test ./...`, sem build tag.
package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const modulo = "github.com/muriloperosa/soat-architecture"

// raizDoRepositorio devolve o diretório raiz do módulo a partir da posição
// deste arquivo (test/architecture/), para o teste não depender do diretório
// de onde o `go test` foi chamado.
func raizDoRepositorio(t *testing.T) string {
	t.Helper()
	_, arquivo, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("não foi possível localizar o arquivo do teste")
	}
	return filepath.Join(filepath.Dir(arquivo), "..", "..")
}

// importsProibidos percorre os .go de produção (sem _test.go nem mocks) em
// dir e devolve "arquivo: import" para cada import que permitido rejeita.
func importsProibidos(t *testing.T, dir string, permitido func(string) bool) []string {
	t.Helper()
	raiz := raizDoRepositorio(t)
	var violacoes []string

	err := filepath.WalkDir(filepath.Join(raiz, dir), func(caminho string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "mocks" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(caminho, ".go") || strings.HasSuffix(caminho, "_test.go") {
			return nil
		}

		arquivo, err := parser.ParseFile(token.NewFileSet(), caminho, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range arquivo.Imports {
			pacote, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if !permitido(pacote) {
				relativo, _ := filepath.Rel(raiz, caminho)
				violacoes = append(violacoes, relativo+": "+pacote)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("erro ao ler %s: %v", dir, err)
	}
	return violacoes
}

// bibliotecaPadrao: na biblioteca padrão do Go o primeiro elemento do caminho
// não tem ponto ("context", "net/http"); em módulos externos tem
// ("gorm.io/gorm", "github.com/...").
func bibliotecaPadrao(pacote string) bool {
	primeiro, _, _ := strings.Cut(pacote, "/")
	return !strings.Contains(primeiro, ".")
}

func dentroDe(pacote string, camadas ...string) bool {
	for _, camada := range camadas {
		prefixo := modulo + "/internal/" + camada
		if pacote == prefixo || strings.HasPrefix(pacote, prefixo+"/") {
			return true
		}
	}
	return false
}

// excecoesDoDominio são as únicas bibliotecas externas que o domínio pode
// importar. Cada entrada precisa de justificativa na ADR 0007.
var excecoesDoDominio = map[string]bool{
	// Hash de senha no VO shared.SenhaHash: algoritmo puro, sem I/O,
	// mantido pelo time do Go.
	"golang.org/x/crypto/bcrypt": true,
}

func TestDominio_DependeSoDaBibliotecaPadraoEDoProprioDominio(t *testing.T) {
	violacoes := importsProibidos(t, "internal/domain", func(pacote string) bool {
		return bibliotecaPadrao(pacote) || dentroDe(pacote, "domain") || excecoesDoDominio[pacote]
	})

	if len(violacoes) > 0 {
		t.Fatalf("o domínio não pode importar aplicação, infraestrutura nem bibliotecas externas fora de excecoesDoDominio:\n  %s",
			strings.Join(violacoes, "\n  "))
	}
}

func TestAplicacao_DependeSoDaBibliotecaPadraoEDoDominio(t *testing.T) {
	violacoes := importsProibidos(t, "internal/application", func(pacote string) bool {
		return bibliotecaPadrao(pacote) || dentroDe(pacote, "domain", "application")
	})

	if len(violacoes) > 0 {
		t.Fatalf("a aplicação não pode importar infraestrutura nem bibliotecas externas (Gin, GORM, JWT...):\n  %s",
			strings.Join(violacoes, "\n  "))
	}
}
