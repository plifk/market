package services

import (
	"crypto/rand"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/plifk/market/config"
)

// Core services include settings of the application, external services, and dependencies.
type Core struct {
	// Settings of the application.
	Settings config.Settings

	// PostgreSQL relational database.
	Postgres *pgxpool.Pool

	// Redis cache layer.
	Redis *redis.Client

	// Elasticsearch client.
	Elasticsearch *elasticsearch.Client

	// CSRFProtection middleware.
	CSRFProtection *CSRFProtection
}

// NewModules creates an instance of each service in this package and returns a Module object that can be injected elsewhere.
func NewModules(core *Core) (*Modules, error) {
	return &Modules{
		Settings: core.Settings,
		Accounts: Accounts{core: core},
		Sessions: Sessions{core: core},
		Security: Security{csrfProtection: core.CSRFProtection},
		Images:   Images{core: core},
	}, nil
}

// Modules exposes internal services to the HTTP handlers without giving direct unchecked access to the core services.
type Modules struct {
	Settings config.Settings
	Accounts Accounts
	Sessions Sessions
	Security Security
	Images   Images
}

func new11RandomID() string {
	const (
		alphabet = "123456789ABCDEFGHJKLMNPQRSTUWVXYZabcdefghijkmnopqrstuwvxyz" // base58
		size     = 11
	)

	var id = make([]byte, size)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	for i, p := range id {
		id[i] = alphabet[int(p)%len(alphabet)] // discard everything but the least significant bits
	}
	return string(id)
}
