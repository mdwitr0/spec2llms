package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Source         string `json:"source"`
	Output         string `json:"output"`
	BaseURL        string `json:"baseUrl"`
	DocsBaseURL    string `json:"docsBaseUrl"`    // базовый URL для ссылок на документацию (llms.txt)
	Title          string `json:"title"`
	Language       string `json:"language"`
	GroupBy        string `json:"groupBy"`        // tag, path
	SkipValidation bool   `json:"skipValidation"` // пропустить валидацию OpenAPI
}

func DefaultConfig() *Config {
	return &Config{
		Output:   "./llms",
		Language: "en",
		GroupBy:  "tag",
	}
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Source == "" {
		return ErrSourceRequired
	}
	return nil
}
