package orm

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleter_Build(t *testing.T) {
	db := memoryDB(t)
	testCase := []struct {
		name      string
		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from ",
			builder: NewDeleter[TestModel](db),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			name:    "empty from ",
			builder: NewDeleter[TestModel](db).From(""),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			name:    "from",
			builder: NewDeleter[TestModel](db).From("`test_Model`"),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_Model`;",
			},
		},
		{
			name:    "from db",
			builder: NewDeleter[TestModel](db).From("`test_db`.`test_Model`"),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_db`.`test_Model`;",
			},
		},
		{
			name:    "where",
			builder: NewDeleter[TestModel](db).Where(C("Age").EQ(18)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `age` = ?;",
				Args: []any{18},
			},
		},
		{
			name:    "not where",
			builder: NewDeleter[TestModel](db).Where(Not(C("Age").EQ(18))),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE  NOT (`age` = ?);",
				Args: []any{18},
			},
		},
		{
			name:    "and where",
			builder: NewDeleter[TestModel](db).Where(C("Age").EQ(18).And(C("FirstName").EQ("Tom"))),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "or where",
			builder: NewDeleter[TestModel](db).Where(C("Age").EQ(18).Or(C("FirstName").EQ("Tom"))),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE (`age` = ?) OR (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "empty where",
			builder: NewDeleter[TestModel](db).Where(),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			name:    "err basic type",
			builder: NewDeleter[int](db).Where(),
			wantErr: errors.New("Model: 不支持类型 int"),
		},
		{
			name:    "err field",
			builder: NewDeleter[TestModel](db).Where(Not(C("XXX").EQ(18))),
			wantErr: errors.New("orm: 未知字段 XXX"),
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}
