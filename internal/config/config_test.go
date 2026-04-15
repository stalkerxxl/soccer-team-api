package config

import (
	"testing"
	"time"
)

func TestConfigValidateAcceptsSupportedAppEnvs(t *testing.T) {
	t.Parallel()

	appEnvs := []string{AppEnvLocal, AppEnvTest, AppEnvProd}
	for _, appEnv := range appEnvs {
		t.Run(appEnv, func(t *testing.T) {
			t.Parallel()

			cfg := Config{
				AppEnv: appEnv,
				Auth: AuthConfig{
					AccessSecret:   "secret",
					AccessTTL:      time.Second,
					PasswordMinLen: 5,
					PasswordMaxLen: 72,
				},
			}

			if err := cfg.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestConfigValidateRejectsUnsupportedAppEnv(t *testing.T) {
	t.Parallel()

	cfg := Config{
		AppEnv: "staging",
		Auth: AuthConfig{
			AccessSecret:   "secret",
			AccessTTL:      time.Second,
			PasswordMinLen: 5,
			PasswordMaxLen: 72,
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}
