package confpostgres

import (
	"github.com/go-courier/sqlx/v2/migration"
	"github.com/kunlun-qilian/confpostgres/tests"
	"testing"
)

func TestPostgres_Connect(t *testing.T) {
	m := &Postgres{
		Host:     "127.0.0.1",
		User:     "postgres",
		Port:     35432,
		DBName:   "example",
		Password: "123456",
		Database: tests.DB,
	}

	m.SetDefaults()
	m.Init()

	if err := migration.Migrate(m.DB, nil); err != nil {
		panic(err)
	}
}
