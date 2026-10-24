package internal

import (
	"sync"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"error"`
	Omada    struct {
		OmadaURL     string `env:"OMADA_URL,required"`
		SiteName     string `env:"OMADA_SITE_NAME,required"`
		ClientID     string `env:"OMADA_CLIENT_ID,required"`
		ClientSecret string `env:"OMADA_CLIENT_SECRET,required"`
		Username     string `env:"OMADA_USERNAME,required"`
		Password     string `env:"OMADA_PASSWORD,required"`
	}
	Prometheus struct {
		MetricsPath string `env:"METRICS_PATH" envDefault:"/metrics"`
		// MetricsHost is the listen address. Empty means all interfaces.
		MetricsHost string `env:"METRICS_HOST" envDefault:""`
		MetricsPort string `env:"METRICS_PORT" envDefault:"8080"`
		// EnableClientMetrics exposes per-client series. It is opt-in because the
		// client list costs one extra API request per scrape and produces one
		// series set per associated client.
		EnableClientMetrics bool `env:"ENABLE_CLIENT_METRICS" envDefault:"false"`
	}
	Loki struct {
		LokiURL     string `env:"LOKI_URL"`
		Environment string `env:"LOKI_ENV" envDefault:"dev"`
		GoVersion   string
		AppName     string
		AppVersion  string `env:"LOKI_APP_VERSION" envDefault:"0.0.0"`
	}
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		var err error
		instance, err = loadConfig()
		if err != nil {
			panic("Failed to load configuration: " + err.Error())
		}
	})
	return instance
}

func loadConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	cfg.Loki.AppName = AppName
	cfg.Loki.GoVersion = goVersion
	return cfg, nil
}
