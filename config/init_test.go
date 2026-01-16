package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestDefaultsSetEnvProd(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })
	viper.Reset()

	setDefaults()

	if got := viper.GetString("env"); got != "prod" {
		t.Fatalf("expected env default prod, got %q", got)
	}
}
