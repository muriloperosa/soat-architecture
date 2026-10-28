package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServe_EncerraGraciosamenteAoCancelarContexto(t *testing.T) {
	srv := &http.Server{Addr: "127.0.0.1:0", Handler: http.NotFoundHandler(), ReadHeaderTimeout: time.Second}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- serve(ctx, srv, time.Second) }()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("serve não retornou após o cancelamento do contexto")
	}
}

func TestServe_RetornaErroQuandoPortaInvalida(t *testing.T) {
	srv := &http.Server{Addr: "127.0.0.1:-1", Handler: http.NotFoundHandler(), ReadHeaderTimeout: time.Second}

	err := serve(context.Background(), srv, time.Second)

	require.Error(t, err)
}
