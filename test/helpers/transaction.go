package helpers

import "context"

// TransactionRunnerMock is a no-op shared.TransactionRunner for use case
// unit tests whose repositories are mocked (there's no real connection to
// open a transaction on). Calls counts how many times Executar ran, so a
// test can assert the use case actually delegates to it instead of just
// running the business logic directly — a plain no-op fake can't tell
// those two cases apart.
type TransactionRunnerMock struct {
	Calls int
}

func (r *TransactionRunnerMock) Executar(ctx context.Context, fn func(ctx context.Context) error) error {
	r.Calls++
	return fn(ctx)
}
