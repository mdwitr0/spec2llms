package main

import (
	"fmt"
	"os"

	"github.com/mdwit/spec2llms/internal/config"
	"github.com/mdwit/spec2llms/internal/generator"
	"github.com/mdwit/spec2llms/internal/parser"
	"github.com/spf13/cobra"
)

var (
	version = "dev"

	cfgFile        string
	output         string
	title          string
	baseURL        string
	language       string
	skipValidation bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "spec2llms [source]",
		Short:   "Generate llms.txt from OpenAPI specification",
		Long:    `spec2llms generates llms.txt files from OpenAPI 3.x specifications for LLM agents.`,
		Version: version,
		Args:    cobra.MaximumNArgs(1),
		RunE:    run,
	}

	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file (spec2llms.json)")
	rootCmd.Flags().StringVarP(&output, "output", "o", "./llms", "output directory")
	rootCmd.Flags().StringVarP(&title, "title", "t", "", "API title")
	rootCmd.Flags().StringVarP(&baseURL, "base-url", "b", "", "base URL for API")
	rootCmd.Flags().StringVarP(&language, "lang", "l", "en", "output language (en, ru)")
	rootCmd.Flags().BoolVar(&skipValidation, "skip-validation", false, "skip OpenAPI spec validation")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig(args)
	if err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	fmt.Printf("Parsing OpenAPI spec: %s\n", cfg.Source)
	api, err := parser.Parse(cfg.Source, &parser.ParseOptions{
		SkipValidation: cfg.SkipValidation,
	})
	if err != nil {
		return fmt.Errorf("failed to parse spec: %w", err)
	}

	fmt.Printf("Found %d endpoints\n", len(api.Endpoints))

	gen := generator.New(cfg, api)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate: %w", err)
	}

	fmt.Printf("Generated llms.txt in %s\n", cfg.Output)
	return nil
}

func loadConfig(args []string) (*config.Config, error) {
	var cfg *config.Config
	var err error

	if cfgFile != "" {
		cfg, err = config.LoadFromFile(cfgFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		cfg = config.DefaultConfig()
	}

	// CLI флаги переопределяют конфиг
	if len(args) > 0 {
		cfg.Source = args[0]
	}
	if output != "" && output != "./llms" {
		cfg.Output = output
	}
	if title != "" {
		cfg.Title = title
	}
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if language != "" {
		cfg.Language = language
	}
	if skipValidation {
		cfg.SkipValidation = true
	}

	return cfg, nil
}
