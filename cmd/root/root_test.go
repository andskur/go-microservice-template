package root

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"microservice-template/config"
	"microservice-template/internal"
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

func TestCmdSetsVersionTemplate(t *testing.T) {
	app, err := internal.NewApplication()
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	cmd := Cmd(app)
	if cmd.VersionTemplate() == "" {
		t.Fatalf("expected version template to be set")
	}
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

	t.Setenv("ENV", "prod-env")

	if err := initializeConfig(cmd, cfg); err != nil {
		t.Fatalf("initializeConfig returned error: %v", err)
	}

	if cfg.Env != "prod-env" {
		t.Fatalf("expected env from ENV, got %q", cfg.Env)
	}

	// Reset and ensure flag wins over env
	resetViper(t)
	cfg.Env = ""
	cmd = newTestCmd()

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

func TestBindFlagsSetsDefaultsFromViper(t *testing.T) {
	resetViper(t)
	t.Cleanup(func() { resetViper(t) })

	cmd := newTestCmd()
	viper.Set("env", "from-viper")

	bindFlags(cmd)

	val, err := cmd.Flags().GetString("env")
	if err != nil {
		t.Fatalf("unexpected error getting flag: %v", err)
	}
	if val != "from-viper" {
		t.Fatalf("expected flag to be set from viper, got %q", val)
	}
}

func writeConfigFile(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "config.yaml")
	content := "env: file-env\nhttp:\n  port: 1234\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestInitializeConfig_ReadsExplicitConfigFileAndEnvOverrides(t *testing.T) {
	resetViper(t)
	t.Cleanup(func() { resetViper(t) })
	viper.SetDefault("http.port", 8080)

	setConfigFile(writeConfigFile(t, t.TempDir()), "test")

	cfg := &config.Scheme{}
	if err := initializeConfig(newTestCmd(), cfg); err != nil {
		t.Fatalf("initializeConfig returned error: %v", err)
	}
	if cfg.Env != "file-env" {
		t.Fatalf("expected env from config file, got %q", cfg.Env)
	}
	if cfg.HTTP == nil || cfg.HTTP.Port != 1234 {
		t.Fatalf("expected http.port 1234 from config file, got %+v", cfg.HTTP)
	}

	// Env vars override the config file.
	t.Setenv("HTTP_PORT", "4321")
	cfg = &config.Scheme{}
	if err := initializeConfig(newTestCmd(), cfg); err != nil {
		t.Fatalf("initializeConfig returned error: %v", err)
	}
	if cfg.HTTP == nil || cfg.HTTP.Port != 4321 {
		t.Fatalf("expected http.port 4321 from env, got %+v", cfg.HTTP)
	}
}

func TestInitializeConfig_FindsConfigInWorkingDir(t *testing.T) {
	resetViper(t)
	t.Cleanup(func() { resetViper(t) })

	dir := t.TempDir()
	writeConfigFile(t, dir)
	t.Chdir(dir)

	setConfigFile("", "test")

	cfg := &config.Scheme{}
	if err := initializeConfig(newTestCmd(), cfg); err != nil {
		t.Fatalf("initializeConfig returned error: %v", err)
	}
	if cfg.Env != "file-env" {
		t.Fatalf("expected env from ./config.yaml, got %q", cfg.Env)
	}
}

func TestInitializeConfig_MissingExplicitConfigFileFails(t *testing.T) {
	resetViper(t)
	t.Cleanup(func() { resetViper(t) })

	setConfigFile(filepath.Join(t.TempDir(), "missing.yaml"), "test")

	if err := initializeConfig(newTestCmd(), &config.Scheme{}); err == nil {
		t.Fatalf("expected error for missing explicit config file")
	}
}
