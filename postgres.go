package confpostgres

import (
	"context"
	"fmt"
	"github.com/kunlun-qilian/sqlx/v3/migration"
	"time"

	"github.com/kunlun-qilian/sqlx/v3/postgresqlconnector"

	"github.com/go-courier/envconf"
	"github.com/kunlun-qilian/sqlx/v3"
	"github.com/spf13/cobra"
)

type Postgres struct {
	DBName          string           `env:""`
	Host            string           `env:",upstream"`
	SlaveHost       string           `env:",upstream"`
	Port            int              `env:""`
	User            string           `env:""`
	Password        envconf.Password `env:""`
	Extra           string
	Extensions      []string
	PoolSize        int
	ConnMaxLifetime envconf.Duration
	Database        *sqlx.Database `env:"-"`

	*sqlx.DB `env:"-"`
	slaveDB  *sqlx.DB `env:"-"`

	commands []*cobra.Command
}

func (m *Postgres) LivenessCheck() map[string]string {
	s := map[string]string{}

	_, err := m.DB.ExecContext(context.Background(), "SELECT 1")
	if err != nil {
		s[m.Host] = err.Error()
	} else {
		s[m.Host] = "ok"
	}

	if m.slaveDB != nil {
		_, err := m.slaveDB.ExecContext(context.Background(), "SELECT 1")
		if err != nil {
			s[m.SlaveHost] = err.Error()
		} else {
			s[m.SlaveHost] = "ok"
		}
	}

	return s
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
}

func (m *Postgres) url(host string) string {
	password := m.Password
	if password != "" {
		password = ":" + password
	}
	return fmt.Sprintf("postgres://%s%s@%s:%d", m.User, password, host, m.Port)
}

func (m *Postgres) conn(host string) (*sqlx.DB, error) {
	db := m.Database.OpenDB(&postgresqlconnector.PostgreSQLConnector{
		Host:       m.url(host),
		Extra:      m.Extra,
		Extensions: m.Extensions,
	})

	db.SetMaxOpenConns(m.PoolSize)
	db.SetMaxIdleConns(m.PoolSize / 2)
	db.SetConnMaxLifetime(time.Duration(m.ConnMaxLifetime))

	_, err := db.ExecContext(context.Background(), "SELECT 1")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (m *Postgres) UseSlave() sqlx.DBExecutor {
	if m.slaveDB != nil {
		return m.slaveDB
	}
	return m.DB
}

func (m *Postgres) Init() {
	// add migrate
	m.commands = append(m.commands, &cobra.Command{
		Use: "migrate",
		Run: func(cmd *cobra.Command, args []string) {
			if err := migration.Migrate(m.DB, nil); err != nil {
				panic(err)
			}
		},
	})

	r := Retry{Repeats: 5, Interval: envconf.Duration(1 * time.Second)}

	err := r.Do(func() error {
		db, err := m.conn(m.Host)
		if err != nil {
			return err
		}
		m.DB = db
		return nil
	})

	if err != nil {
		panic(err)
	}

	if m.SlaveHost != "" {
		err := r.Do(func() error {
			db, err := m.conn(m.SlaveHost)
			if err != nil {
				return err
			}
			m.slaveDB = db
			return nil
		})

		if err != nil {
			panic(err)
		}
	}
}

func SwitchSlave(executor sqlx.DBExecutor) sqlx.DBExecutor {
	if canSlave, ok := executor.(CanSlave); ok {
		return canSlave.UseSlave()
	}
	return executor
}

type CanSlave interface {
	UseSlave() sqlx.DBExecutor
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
