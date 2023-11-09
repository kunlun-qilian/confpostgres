package confpostgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-courier/envconf"
	"github.com/kunlun-qilian/sqlx/v3"
	"github.com/kunlun-qilian/sqlx/v3/postgresqlconnector"
)

type PostgresEndpoint struct {
	Endpoint      envconf.Endpoint `env:""`
	SlaveEndpoint envconf.Endpoint `env:""`
	DBName        string           `env:",opt"`
	Database      *sqlx.Database   `env:"-"`

	Extensions      []string
	PoolSize        int
	ConnMaxLifetime envconf.Duration

	*sqlx.DB `env:"-"`
	slaveDB  *sqlx.DB `env:"-"`
}

func (m *PostgresEndpoint) LivenessCheck() map[string]string {
	s := map[string]string{}

	_, err := m.DB.ExecContext(context.Background(), "SELECT 1")
	if err != nil {
		s[m.Endpoint.Host()] = err.Error()
	} else {
		s[m.Endpoint.Host()] = "ok"
	}

	if m.slaveDB != nil {
		_, err := m.slaveDB.ExecContext(context.Background(), "SELECT 1")
		if err != nil {
			s[m.SlaveEndpoint.Host()] = err.Error()
		} else {
			s[m.SlaveEndpoint.Host()] = "ok"
		}
	}

	return s
}

func (m *PostgresEndpoint) SetDefaults() {
	if m.PoolSize == 0 {
		m.PoolSize = 10
	}

	if m.ConnMaxLifetime == 0 {
		m.ConnMaxLifetime = envconf.Duration(1 * time.Hour)
	}

	if m.Endpoint.IsZero() {
		m.Endpoint.Hostname = "127.0.0.1"
		m.Endpoint.Port = 5432
	}

	if m.Database.Name == "" {
		if len(m.Endpoint.Base) > 0 {
			m.Database.Name = m.Endpoint.Base
		}
		if m.DBName != "" {
			m.Database.Name = m.DBName
		}
	}
}

func (m *PostgresEndpoint) url(host string) string {
	password := m.Endpoint.Password
	if password != "" {
		password = ":" + password
	}
	return fmt.Sprintf("postgres://%s%s@%s", m.Endpoint.Username, password, host)
}

func (m *PostgresEndpoint) conn(host string, readonly bool) (*sqlx.DB, error) {
	connector := &postgresqlconnector.PostgreSQLConnector{
		Host:  m.url(host),
		Extra: m.Endpoint.Extra.Encode(),
	}
	if !readonly {
		connector.Extensions = m.Extensions
	}
	db := m.Database.OpenDB(connector)

	db.SetMaxOpenConns(m.PoolSize)
	db.SetMaxIdleConns(m.PoolSize / 2)
	db.SetConnMaxLifetime(time.Duration(m.ConnMaxLifetime))

	_, err := db.ExecContext(context.Background(), "SELECT 1")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (m *PostgresEndpoint) UseSlave() sqlx.DBExecutor {
	if m.slaveDB != nil {
		return m.slaveDB
	}
	return m.DB
}

func (m *PostgresEndpoint) Init() {
	// 执行 dockerize 时不连接数据库
	if len(os.Args) > 1 && os.Args[1] == "dockerize" {
		return
	}

	// 若配置中指定库名，覆盖默认值
	if len(m.Endpoint.Base) > 0 {
		m.Database.Name = m.Endpoint.Base
	}
	r := Retry{Repeats: 5, Interval: envconf.Duration(1 * time.Second)}

	err := r.Do(func() error {
		db, err := m.conn(m.Endpoint.Host(), false)
		if err != nil {
			return err
		}
		m.DB = db
		return nil
	})

	if err != nil {
		panic(err)
	}

	if !m.SlaveEndpoint.IsZero() {
		err := r.Do(func() error {
			db, err := m.conn(m.SlaveEndpoint.Host(), true)
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
