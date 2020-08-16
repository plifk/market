package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Settings loaded from the configuration file.
type Settings struct {
	// Debug mode useful to print .
	Debug bool

	// HTTPAddress to receive incoming HTTP requests.
	HTTPAddress string

	// HTTPCertFile path to a TLS certificate. Set this to listen for HTTPS requests.
	HTTPCertFile string

	// HTTPKeyFile path to a certificate key. Set this to listen for HTTPS requests.
	HTTPKeyFile string

	// SQLDataSourceName for the PostgreSQL database.
	SQLDataSourceName string

	// RedisAddress for the key-value storage.
	RedisAddress string

	// RedisUsername for the key-value storage.
	RedisUsername string

	// RedisPassword for the key-value storage.
	RedisPassword string

	// ElasticsearchHost for the search engine.
	ElasticsearchHost string

	// ElasticsearchUsername for the search engine.
	ElasticsearchUsername string

	// ElasticsearchPassword for the search engine.
	ElasticsearchPassword string

	// ElasticsearchAPIKey for the search engine.
	ElasticsearchAPIKey string

	// FileStorageHost to a storage service such as AWS S3, MinIO, or Google Storage.
	FileStorageHost string

	// FileStorageKey of the storage service.
	FileStorageAPIKey string

	// FileStorageSecret of the storage service.
	FileStorageSecret string

	// HTTPInspectionAddress for Go pprof and expvar. Only accessible on localhost.
	HTTPInspectionAddress string

	// StaticDirectory where regular files are stored.
	// TODO(henvic): Embed files in the binary (see https://github.com/golang/go/issues/35950).
	StaticDirectory string

	// TemplatesDirectory where HTML templates are stored.
	TemplatesDirectory string

	// ThumbnailServiceHost for the imaginary microservice.
	ThumbnailServiceHost string
}

// ReadFile loads the settings from a configuration file.
func ReadFile(path string) (s Settings, err error) {
	f, err := os.Open(path) // #nosec
	if err != nil {
		return s, err
	}
	if err = json.NewDecoder(f).Decode(&s); err != nil {
		return s, fmt.Errorf("cannot load market configuration: %w", err)
	}
	return s, nil
}
