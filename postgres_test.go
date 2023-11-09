package confpostgres

import (
	"context"
	"testing"

	"github.com/onsi/gomega"

	"github.com/kunlun-qilian/sqlx/v3"
)

func Test(t *testing.T) {

	pg := &Postgres{
		Host:      "127.0.0.1",
		SlaveHost: "127.0.0.1",
		User:      "postgres",
		Database: &sqlx.Database{
			Name: "osm",
		},
		Extensions: []string{
			"postgis", "hstore",
		},
	}

	pg.SetDefaults()
	pg.Init()

	{
		row, err := pg.QueryContext(context.Background(), "SELECT 1")
		gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
		row.Close()
	}

	row, err := SwitchSlave(pg).QueryContext(context.Background(), "SELECT 1")
	gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
	row.Close()

	gomega.NewWithT(t).Expect(pg.UseSlave()).NotTo(gomega.Equal(pg.DB))

	for i := 0; i < 100; i++ {
		t.Log(pg.LivenessCheck())
	}
}
