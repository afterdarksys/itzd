package config

import (
	"os"
	"path/filepath"
	"testing"
)

func isolateConfig(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("ITZ_TOKEN", "")
	t.Setenv("ITZ_API", "")
	t.Setenv("ITZ_VALIDATOR_ENDPOINT", "")
	t.Setenv("ITZ_VALIDATOR_SERVER_NAME", "")
	t.Setenv("ITZ_DEFAULT_ZONE", "")
	return root
}

func TestConfigUsesXDGAndEnvironmentPrecedence(t *testing.T) {
	root := isolateConfig(t)
	if err := Save(&Config{APIEndpoint: "https://disk.example", Token: "disk-token"}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ITZ_API", "https://env.example")
	t.Setenv("ITZ_TOKEN", "env-token")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIEndpoint != "https://env.example" || cfg.Token != "env-token" {
		t.Fatalf("unexpected config %+v", cfg)
	}
	want := filepath.Join(root, "itzd", configFileName)
	if got, _ := path(); got != want {
		t.Fatalf("path=%q want=%q", got, want)
	}
}

func TestConfigSaveAlwaysUses0600(t *testing.T) {
	isolateConfig(t)
	if err := Save(&Config{Token: "secret"}); err != nil {
		t.Fatal(err)
	}
	p, _ := path()
	if err := os.Chmod(p, 0644); err != nil {
		t.Fatal(err)
	}
	if err := Save(&Config{Token: "secret-2"}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("mode=%o", info.Mode().Perm())
	}
}

func TestConfigInvalidJSONFailsClosed(t *testing.T) {
	isolateConfig(t)
	p, _ := path()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
