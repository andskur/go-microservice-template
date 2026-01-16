package root

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"microservice-template/config"
)

func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	applyDefaults()
}

func applyDefaults() {
	viper.SetDefault("env", "prod")
}

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("env", "", "environment")
	return cmd
}

func TestInitializeConfig_IgnoresMissingFile(t *testing.T) {
	resetViper(t)
	t.Cleanup(func() { resetViper(t) })

	cfg := &config.Scheme{}
	cmd := newTestCmd()

	// Point to a non-existent file but ensure path exists; viper.ReadInConfig
	// treats missing file as ConfigFileNotFoundError only when a config name is set.
	tmp := t.TempDir()
	viper.AddConfigPath(tmp)
	viper.SetConfigName("nope")

	if err := initializeConfig(cmd, cfg); err != nil {
		t.Fatalf("expected missing config file to be ignored, got: %v", err)
	}
}

func TestInitializeConfig_PropagatesReadError(t *testing.T) {
	t.Cleanup(func() { resetViper(t) })

	cfg := &config.Scheme{}
	cmd := newTestCmd()

	// force a read error by pointing to a directory
	dir := t.TempDir()
	viper.SetConfigFile(dir) // not a file

	err := initializeConfig(cmd, cfg)
	if err == nil {
		t.Fatalf("expected error for invalid config path")
	}
	if !errors.Is(err, os.ErrInvalid) && !errors.Is(err, os.ErrPermission) && !errors.Is(err, os.ErrNotExist) {
		// viper may wrap a variety of underlying errors; ensure it's not a nil or ignored case
		// just assert it's not the ConfigFileNotFound path by type check
		var cfgNotFound viper.ConfigFileNotFoundError
		if errors.As(err, &cfgNotFound) {
			t.Fatalf("expected real error, got ConfigFileNotFoundError: %v", err)
		}
	}
}

func TestInitializeConfig_BindsEnvAndFlags(t *testing.T) {
	resetViper(t)
	t.Cleanup(func() { resetViper(t) })

	cfg := &config.Scheme{}
	cmd := newTestCmd()

	// env should set when flag not provided
	t.Setenv("ENV", "prod-env")

	// bind flags inside initializeConfig
	if err := initializeConfig(cmd, cfg); err != nil {
		t.Fatalf("initializeConfig returned error: %v", err)
	}

	if cfg.Env != "prod-env" {
		t.Fatalf("expected env from ENV, got %q", cfg.Env)
	}

	// Ensure viper state reset before second call
	resetViper(t)
	cfg.Env = ""
	cmd = newTestCmd()

	// If flag is set, it should override env
	if err := cmd.Flags().Set("env", "flag-env"); err != nil {
		t.Fatalf("failed setting flag: %v", err)
	}

	t.Setenv("ENV", "env-ignored")

	if err := initializeConfig(cmd, cfg); err != nil {
		t.Fatalf("initializeConfig returned error: %v", err)
	}

	if cfg.Env != "flag-env" {
		t.Fatalf("expected env from flag, got %q", cfg.Env)
	}
}
