package confpostgres

import (
	"context"
	"testing"

	"github.com/onsi/gomega"

	"github.com/kunlun-qilian/sqlx/v3"
)

func TestEndpoint(t *testing.T) {

	pg := &PostgresEndpoint{
		Database: &sqlx.Database{
			Name: "osm",
		},
	}

	_ = pg.Endpoint.UnmarshalText([]byte("postgresql://postgres:root@127.0.0.1:5432/test"))
	_ = pg.SlaveEndpoint.UnmarshalText([]byte("postgresql://postgres:root@127.0.0.1:5432"))
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
