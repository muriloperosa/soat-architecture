package main

import (
	"context"
	"net"
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

func TestServe_TerminaRequisicaoEmAndamentoAntesDeEncerrar(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	iniciou := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(iniciou)
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: time.Second}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- serve(ctx, srv, 2*time.Second) }()

	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, time.Second, 10*time.Millisecond, "servidor não começou a escutar")

	type resultado struct {
		status int
		err    error
	}
	res := make(chan resultado, 1)
	go func() {
		resp, err := http.Get("http://" + addr)
		if err != nil {
			res <- resultado{err: err}
			return
		}
		_ = resp.Body.Close()
		res <- resultado{status: resp.StatusCode}
	}()

	<-iniciou
	cancel()

	r := <-res
	require.NoError(t, r.err, "requisição em andamento foi cortada no encerramento")
	require.Equal(t, http.StatusOK, r.status)
	require.NoError(t, <-done)
}
