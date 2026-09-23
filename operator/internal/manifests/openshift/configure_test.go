package openshift

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/loki/operator/internal/manifests/internal/config"
)

func TestAlertGeneratorURL(t *testing.T) {
	require.Equal(t, "", AlertGeneratorURL(""))
	require.Equal(t, "https://console.apps.example.com/monitoring/logs", AlertGeneratorURL("https://console.apps.example.com"))
	require.Equal(t, "https://console.apps.example.com/monitoring/logs", AlertGeneratorURL("https://console.apps.example.com/"))
}

func TestConfigureDefaultMonitoringAM_SetsConsoleExternalURL(t *testing.T) {
	cfg := &config.Options{}
	err := configureDefaultMonitoringAM(cfg, "https://console.apps.example.com")
	require.NoError(t, err)
	require.Equal(t, "https://console.apps.example.com/monitoring/logs", cfg.Ruler.AlertManager.ExternalURL)
}

func TestConfigureDefaultMonitoringAM_KeepsUserExternalURL(t *testing.T) {
	cfg := &config.Options{
		Ruler: config.Ruler{
			AlertManager: &config.AlertManagerConfig{
				ExternalURL: "https://grafana.example.com",
			},
		},
	}
	err := configureDefaultMonitoringAM(cfg, "https://console.apps.example.com")
	require.NoError(t, err)
	require.Equal(t, "https://grafana.example.com", cfg.Ruler.AlertManager.ExternalURL)
}
