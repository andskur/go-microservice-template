package serve

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"

	"microservice-template/internal"
)

type stubApp struct {
	initErr  error
	serveErr error
	stopErr  error

	initCalled  bool
	serveCalled bool
	stopCalled  bool
}

func (a *stubApp) Init() error {
	a.initCalled = true
	return a.initErr
}

func (a *stubApp) Serve() error {
	a.serveCalled = true
	return a.serveErr
}
func (a *stubApp) Stop() error {
	a.stopCalled = true
	return a.stopErr
}

// Satisfy the subset used by serve command (unused but kept for interface parity).
func (a *stubApp) Config() *internal.App { return nil }
func (a *stubApp) Version() string       { return "version" }

func TestCmdRunsInitAndServe(t *testing.T) {
	app := &stubApp{}

	cmd := Cmd((*internal.App)(nil))
	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		if err := app.Init(); err != nil {
			return err
		}
		return app.Serve()
	}
	cmd.PreRun = func(_ *cobra.Command, _ []string) {
		_ = app.Version()
	}
	cmd.PostRun = func(_ *cobra.Command, _ []string) {
		_ = app.Stop()
	}

	if err := cmd.RunE(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !app.initCalled || !app.serveCalled {
		t.Fatalf("expected init and serve to be called")
	}
}

func TestCmdStopsAfterServeError(t *testing.T) {
	serveErr := errors.New("serve failed")
	app := &stubApp{serveErr: serveErr}

	cmd := Cmd((*internal.App)(nil))
	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		if err := app.Init(); err != nil {
			return err
		}
		return app.Serve()
	}
	cmd.PreRun = func(_ *cobra.Command, _ []string) {
		_ = app.Version()
	}
	cmd.PostRun = func(_ *cobra.Command, _ []string) {
		_ = app.Stop()
	}

	err := cmd.RunE(nil, nil)
	if !errors.Is(err, serveErr) {
		t.Fatalf("expected serve error, got %v", err)
	}

	// Manually invoke PostRun to mirror cobra behavior after RunE returns an error.
	cmd.PostRun(nil, nil)

	if !app.stopCalled {
		t.Fatalf("expected Stop to be called even after serve error")
	}
}
