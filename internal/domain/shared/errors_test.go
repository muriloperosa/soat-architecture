package shared

import (
	"errors"
	"fmt"
	"testing"
)

func TestAppError_Error_SemErroEncapsulado(t *testing.T) {
	err := NewNotFoundError("cliente não encontrado")

	if got := err.Error(); got != "cliente não encontrado" {
		t.Fatalf("got %q, want %q", got, "cliente não encontrado")
	}
}

func TestAppError_Error_ComErroEncapsulado(t *testing.T) {
	cause := errors.New("connection refused")
	err := NewInternalError("erro ao consultar banco", cause)

	want := "erro ao consultar banco: connection refused"
	if got := err.Error(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAppError_Unwrap_RetornaErroEncapsulado(t *testing.T) {
	cause := errors.New("boom")
	err := NewInternalError("erro interno", cause)

	if !errors.Is(err, cause) {
		t.Fatal("errors.Is deveria encontrar a causa encapsulada")
	}
}

func TestAppError_Unwrap_RetornaNilQuandoSemCausa(t *testing.T) {
	err := NewNotFoundError("cliente não encontrado")

	if err.Unwrap() != nil {
		t.Fatalf("esperava Unwrap nil, obteve %v", err.Unwrap())
	}
}

func TestConstrutores_DefinemKindEDetails(t *testing.T) {
	cases := []struct {
		name string
		err  *AppError
		want ErrorKind
	}{
		{"not found", NewNotFoundError("x"), KindNotFound},
		{"validation", NewValidationError("x"), KindValidation},
		{"conflict", NewConflictError("x"), KindConflict},
		{"internal", NewInternalError("x", nil), KindInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Kind != tc.want {
				t.Fatalf("got Kind %q, want %q", tc.err.Kind, tc.want)
			}
		})
	}
}

func TestNewValidationErrorComDetails_DefineDetails(t *testing.T) {
	err := NewValidationErrorComDetails("dados inválidos", []string{"nome é obrigatório", "email inválido"})

	if err.Kind != KindValidation {
		t.Fatalf("got Kind %q, want %q", err.Kind, KindValidation)
	}
	if len(err.Details) != 2 || err.Details[0] != "nome é obrigatório" {
		t.Fatalf("got Details %v, want detalhes preservados", err.Details)
	}
}

func TestAppError_ErrorsAsFuncionaAtravesDeFmtWrap(t *testing.T) {
	inner := NewNotFoundError("cliente não encontrado")
	wrapped := fmt.Errorf("use case falhou: %w", inner)

	var target *AppError
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As deveria extrair *AppError através do fmt.Errorf %w")
	}
	if target.Kind != KindNotFound {
		t.Fatalf("got Kind %q, want %q", target.Kind, KindNotFound)
	}
}
