package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"autobangumi-go/config"
	"github.com/stretchr/testify/require"
)

func TestLoadExampleConfig(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	cfg, err := config.Load(filepath.Join(pwd, "config.example.toml"))
	require.NoError(t, err)
	require.NotNil(t, cfg)
}
