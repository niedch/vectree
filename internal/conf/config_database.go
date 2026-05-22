package conf

type Database struct {
	ConnectionString string `koanf:"connection_string"`
}
