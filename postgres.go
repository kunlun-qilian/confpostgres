package confpostgres

import (
	"context"
	"fmt"
	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/migration"
	"time"

	"github.com/go-courier/sqlx/v2/postgresqlconnector"

	"github.com/go-courier/envconf"
	"github.com/spf13/cobra"
)

type Postgres struct {
	// Database Name
	DBName          string `env:""`
	Host            string `env:""`
	Port            int
	User            string           `env:""`
	Password        envconf.Password `env:""`
	Extra           string
	Extensions      []string
	PoolSize        int
	ConnMaxLifetime envconf.Duration
	Database        *sqlx.Database `env:"-"`
	*sqlx.DB        `env:"-"`

	commands []*cobra.Command

	Retry
}

func (m *Postgres) SetDefaults() {
	m.Database.Name = m.DBName

	if m.Host == "" {
		m.Host = "127.0.0.1"
	}

	if m.Port == 0 {
		m.Port = 5432
	}

	if m.PoolSize == 0 {
		m.PoolSize = 10
	}

	if m.ConnMaxLifetime == 0 {
		m.ConnMaxLifetime = envconf.Duration(1 * time.Hour)
	}

	if m.Extra == "" {
		m.Extra = "sslmode=disable"
	}

	if m.Repeats == 0 {
		m.Repeats = 3
	}
	if m.Interval == 0 {
		m.Interval = envconf.Duration(10 * time.Second)
	}
}

func (m *Postgres) url(host string) string {
	password := m.Password
	if password != "" {
		password = ":" + password
	}
	return fmt.Sprintf("postgres://%s%s@%s:%d", m.User, password, host, m.Port)
}

func (m *Postgres) Connect() error {
	db := m.Database.OpenDB(&postgresqlconnector.PostgreSQLConnector{
		Host:       m.url(m.Host),
		Extra:      m.Extra,
		Extensions: m.Extensions,
	})

	db.SetMaxOpenConns(m.PoolSize)
	db.SetMaxIdleConns(m.PoolSize / 2)
	db.SetConnMaxLifetime(time.Duration(m.ConnMaxLifetime))

	_, err := db.ExecContext(context.Background(), "SELECT 1")
	if err != nil {
		panic(err)
	}
	m.DB = db

	return nil

}

func (m *Postgres) Init() {

	// migrate
	m.commands = append(m.commands, &cobra.Command{
		Use: "migrate",
		Run: func(cmd *cobra.Command, args []string) {
			if err := migration.Migrate(m.DB, nil); err != nil {
				panic(err)
			}
		},
	})

	if m.DB == nil {
		_ = m.Do(m.Connect)
	}
}

func (m *Postgres) Get() *sqlx.DB {
	if m.DB == nil {
		panic(fmt.Errorf("get db before init"))
	}
	return m.DB
}

func (m *Postgres) Commands() []*cobra.Command {
	return m.commands
}
