package conf

import (
	"log"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Sources   map[string]Source `koanf:"sources"`
	Database  Database          `koanf:"database"`
	Pipeline  Pipeline          `koanf:"pipeline"`
	Chunking  Chunking          `koanf:"chunking"`
	AI        AI                `koanf:"ai"`
	Retrieval Retrieval         `koanf:"retrieval"`
}

func Load() (*Config, error) {
	return loadCustomFile("config.toml")
}

func loadCustomFile(filepath string) (*Config, error) {
	k := koanf.New(".")

	loadDefaults(k)
	loadEnvironment(k)
	loadLocalFile(k, filepath)

	cfg := marshalConf(k)
	err := validateConfig(cfg)

	return cfg, err
}

func marshalConf(k *koanf.Koanf) *Config {
	var cfg Config
	err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"})
	if err != nil {
		log.Fatalf("error unmarshaling config: %v", err)
	}

	for name, src := range cfg.Sources {
		src.Name = name
		cfg.Sources[name] = src
	}

	return &cfg
}

func loadDefaults(k *koanf.Koanf) {
	k.Load(structs.Provider(Config{
		Chunking: Chunking{
			Strategy: MDAST_STRATEGY,
		},
		Pipeline: Pipeline{
			EmbedderBatchSize: 64,
			EmbedderWorkers:   8,
			StoreBatchSize:    8,
			CrawlerWorkers:    4,
		},
		Database: Database{
			ConnectionString: "kownledgebase.db?cache=shared&mode=rw",
		},
	}, "koanf"), nil)
}

func loadEnvironment(k *koanf.Koanf) {
	envCB := func(key string) string {
		switch key {
		case "GEMINI_API_KEY":
			return "ai.gemini_api_key"
		case "OPENAI_API_KEY":
			return "ai.openai_api_key"
		default:
			return key
		}
	}
	k.Load(env.Provider("", ".", envCB), nil)
}

func loadLocalFile(k *koanf.Koanf, filename string) {
	if err := k.Load(file.Provider(filename), toml.Parser()); err != nil {
		log.Println("Cannot load local config file", err)
	}
}
