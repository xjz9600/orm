//go:build e2e

package integration

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"orm"
	"orm/internal/test"
	"testing"
	"time"
)

type InsertSuite struct {
	Suite
}

func TestMySQLInsert(t *testing.T) {
	suite.Run(t, &InsertSuite{
		Suite: Suite{
			driver: "mysql",
			dsn:    "root:root@tcp(localhost:13306)/integration_test",
		},
	})
}

func (i *InsertSuite) TestInsert() {
	db := i.db
	t := i.T()
	testCases := []struct {
		name         string
		i            *orm.Inserter[test.SimpleStruct]
		wantAffected int64
	}{
		{
			name:         "insert one",
			i:            orm.NewInserter[test.SimpleStruct](db).Values(test.NewSimpleStruct(12)),
			wantAffected: 1,
		},
		{
			name: "insert multiple",
			i: orm.NewInserter[test.SimpleStruct](db).Values(test.NewSimpleStruct(13),
				test.NewSimpleStruct(14)),
			wantAffected: 2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			res := tc.i.Exec(ctx)
			affected, err := res.RowsAffected()
			assert.NoError(t, err)
			assert.Equal(t, tc.wantAffected, affected)
		})
	}
}

func (i *InsertSuite) TearDownTest() {
	orm.RawQuery[test.SimpleStruct](i.db, "TRUNCATE TABLE `simple_struct`").Exec(context.Background())
}
