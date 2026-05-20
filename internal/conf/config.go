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
	Sources        map[string]Source `koanf:"sources"`
	Database       Database          `koanf:"database"`
	Pipeline       Pipeline          `koanf:"pipeline"`
	AI             AI                `koanf:"ai"`
	Retrieval      Retrieval         `koanf:"retrieval"`

	GEMINI_API_KEY string            `koanf:"GEMINI_API_KEY"`
}

type Database struct {
	ConnectionString string `koanf:"connection_string"`
}

type Pipeline struct {
	EmbedderBatchSize int    `koanf:"embedder_batch_size"`
	EmbedderWorkers   int    `koanf:"embedder_workers"`
	StoreBatchSize    int    `koanf:"store_batch_size"`
	DocuLoaderWorkers int    `koanf:"docu_loader_workers"`
}

type AI struct {
	EmbeddingModel string `koanf:"embedding_model"`
}

type Retrieval struct {
	SimilarityResults int `koanf:"similarity_results"`
}

func Load() *Config {
	k := koanf.New(".")

	loadDefaults(k)
	loadEnvironment(k)
	loadLocalFile(k)

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
		AI: AI{
			EmbeddingModel: "embedding-001",
		},
		Pipeline: Pipeline{
			EmbedderBatchSize: 64,
			EmbedderWorkers:   8,
			StoreBatchSize:    8,
			DocuLoaderWorkers: 10,
		},
		Database: Database{
			ConnectionString: "kownledgebase.db?cache=shared&mode=rw",
		},
	}, "."), nil)
}

func loadEnvironment(k *koanf.Koanf) {
	k.Load(env.Provider("", ".", nil), nil)
}

func loadLocalFile(k *koanf.Koanf) {
	if err := k.Load(file.Provider("config.toml"), toml.Parser()); err != nil {
		log.Println("Cannot load local config file", err)
	}
}
