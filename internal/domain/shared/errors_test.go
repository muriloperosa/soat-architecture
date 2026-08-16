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
		{"forbidden", NewForbiddenError("x"), KindForbidden},
		{"unauthorized", NewUnauthorizedError("x"), KindUnauthorized},
		{"unavailable", NewUnavailableError("x", nil), KindUnavailable},
		{"internal custom", NewInternalErrorCustom("x"), KindInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Kind != tc.want {
				t.Fatalf("got Kind %q, want %q", tc.err.Kind, tc.want)
			}
		})
	}
}

func TestNewValidationErrorWithDetails_DefineDetails(t *testing.T) {
	err := NewValidationErrorWithDetails("dados inválidos", []string{"nome é obrigatório", "email inválido"})

	if err.Kind != KindValidation {
		t.Fatalf("got Kind %q, want %q", err.Kind, KindValidation)
	}
	if len(err.Details) != 2 || err.Details[0] != "nome é obrigatório" {
		t.Fatalf("got Details %v, want detalhes preservados", err.Details)
	}
}

func TestNewInternalErrorCustom_NaoEncapsulaCausa(t *testing.T) {
	err := NewInternalErrorCustom("falha estatica de infra")

	if err.Kind != KindInternal {
		t.Fatalf("got Kind %q, want %q", err.Kind, KindInternal)
	}
	if err.Message != "falha estatica de infra" {
		t.Fatalf("got Message %q, want a mensagem original", err.Message)
	}
	if err.Err != nil {
		t.Fatalf("got Err %v, want nil (sem causa encapsulada)", err.Err)
	}
	if err.Error() != "falha estatica de infra" {
		t.Fatalf("got Error() %q, want só a mensagem (sem sufixo de causa)", err.Error())
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
