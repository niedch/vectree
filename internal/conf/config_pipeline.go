package conf

type Pipeline struct {
	EmbedderBatchSize int `koanf:"embedder_batch_size"`
	EmbedderWorkers   int `koanf:"embedder_workers"`
	StoreBatchSize    int `koanf:"store_batch_size"`
	CrawlerWorkers    int `koanf:"crawler_workers"`
}
