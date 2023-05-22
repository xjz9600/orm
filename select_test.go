package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"orm/internal/errs"
	"orm/internal/valuer"
	"testing"
)

func TestSelector_Select(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		s         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "multiple columns",
			s:    NewSelector[TestModel](db).Select(C("FirstName"), C("LastName")),
			wantQuery: &Query{
				SQL: "SELECT `first_name`,`last_name` FROM `test_model`;",
			},
		},
		{
			name: "alias columns",
			s:    NewSelector[TestModel](db).Select(C("FirstName").As("my_name")),
			wantQuery: &Query{
				SQL: "SELECT `first_name` AS `my_name` FROM `test_model`;",
			},
		},
		{
			name: "avg",
			s:    NewSelector[TestModel](db).Select(Avg("FirstName")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`first_name`) FROM `test_model`;",
			},
		},
		{
			name: "avg alias",
			s:    NewSelector[TestModel](db).Select(Avg("FirstName").As("my_name")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`first_name`) AS `my_name` FROM `test_model`;",
			},
		},
		{
			name: "raw expression",
			s:    NewSelector[TestModel](db).Select(Raw("COUNT(DISTINCT `first_name`)")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(DISTINCT `first_name`) FROM `test_model`;",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.s.Build()
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, q, tc.wantQuery)

		})

	}
}

func TestSelector_Build(t *testing.T) {
	db := memoryDB(t)
	testCase := []struct {
		name      string
		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from ",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name:    "empty from ",
			builder: NewSelector[TestModel](db).From(""),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name:    "from",
			builder: NewSelector[TestModel](db).From("`test_Model`"),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_Model`;",
			},
		},
		{
			name:    "from db",
			builder: NewSelector[TestModel](db).From("`test_db`.`test_Model`"),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_db`.`test_Model`;",
			},
		},
		{
			name:    "where",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` = ?;",
				Args: []any{18},
			},
		},
		{
			name:    "not where",
			builder: NewSelector[TestModel](db).Where(Not(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`age` = ?);",
				Args: []any{18},
			},
		},
		{
			name:    "and where",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).And(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "or where",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).Or(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) OR (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "raw expression as predicate",
			builder: NewSelector[TestModel](db).Where(Raw("`id` < ?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` < ?;",
				Args: []any{18},
			},
		},
		{
			name:    "raw expression user in predicate",
			builder: NewSelector[TestModel](db).Where(C("Id").Eq(Raw("`age`+?", 1).AsPredicate())),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = (`age`+?);",
				Args: []any{1},
			},
		},
		{
			name:    "columns alias in where",
			builder: NewSelector[TestModel](db).Where(C("Id").As("my_id").Eq(18)),
			wantErr: errs.ErrAliasWhere,
		},
		{
			name:    "empty where",
			builder: NewSelector[TestModel](db).Where(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name:    "group by id",
			builder: NewSelector[TestModel](db).GroupBy(C("Id")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `id`;",
			},
		},
		{
			name:    "group by id firstName",
			builder: NewSelector[TestModel](db).Select(C("FirstName"), C("LastName")).GroupBy(C("FirstName"), C("LastName")),
			wantQuery: &Query{
				SQL: "SELECT `first_name`,`last_name` FROM `test_model` GROUP BY `first_name`,`last_name`;",
			},
		},
		{
			name:    "having",
			builder: NewSelector[TestModel](db).Having(C("Id").Eq(2), C("Age").Eq(12)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` HAVING (`id` = ?) AND (`age` = ?);",
				Args: []any{2, 12},
			},
		},
		{
			name:    "having",
			builder: NewSelector[TestModel](db).Having(Avg("Id").LT(5)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` HAVING AVG(`id`) < ?;",
				Args: []any{5},
			},
		},
		{
			name:    "having and",
			builder: NewSelector[TestModel](db).Having(Avg("Id").LT(5), Avg("FirstName").EQ("Tom")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` HAVING (AVG(`id`) < ?) AND (AVG(`first_name`) = ?);",
				Args: []any{5, "Tom"},
			},
		},
		{
			name:    "column",
			builder: NewSelector[TestModel](db).OrderBy(Asc("Age")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` ORDER BY `age` ASC;",
			},
		},
		{
			name:    "columns",
			builder: NewSelector[TestModel](db).OrderBy(Asc("Age"), Desc("Id")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` ORDER BY `age` ASC,`id` DESC;",
			},
		},
		{
			name:    "offset only",
			builder: NewSelector[TestModel](db).Offset(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` OFFSET ?;",
				Args: []any{10},
			},
		},
		{
			name:    "limit only",
			builder: NewSelector[TestModel](db).Limit(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` LIMIT ?;",
				Args: []any{10},
			},
		},
		{
			name:    "limit offset",
			builder: NewSelector[TestModel](db).Limit(20).Offset(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` LIMIT ? OFFSET ?;",
				Args: []any{20, 10},
			},
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

func TestSelector_GET(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	defer mockDb.Close()
	require.NoError(t, err)
	db, err := OpenDB(mockDb)
	require.NoError(t, err)
	mock.ExpectQuery("SELECT .*").WithArgs(1).WillReturnError(errors.New("query error"))

	mock.ExpectQuery("SELECT .*").WithArgs(1).WillReturnRows(mock.NewRows([]string{"id", "first_name", "age", "last_name"}))

	mockRows := mock.NewRows([]string{"id", "first_name", "age", "last_name"})
	mockRows.AddRow(1, "Tom", 18, "Jerry")
	mock.ExpectQuery("SELECT .*").WithArgs(1).WillReturnRows(mockRows)
	testCase := []struct {
		name    string
		s       *Selector[TestModel]
		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "invalid query",
			s:       NewSelector[TestModel](db).Where(C("xxx").Eq(1)),
			wantErr: errs.NewErrUnKnownField("xxx"),
		},
		{
			name:    "query err",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantErr: errors.New("query error"),
		},
		{
			name:    "no rows",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantErr: ErrNoRows,
		},
		{
			name: "data",
			s:    NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantRes: &TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  sql.NullString{Valid: true, String: "Jerry"},
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			model, err := tc.s.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, model)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  sql.NullString
}

func memoryDB(t *testing.T, opts ...DBOptions) *DB {
	db, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory", opts...)
	require.NoError(t, err)
	return db
}

func BenchmarkQuerier_Get(b *testing.B) {
	db, err := Open("sqlite3", fmt.Sprintf("file:benchmark_get.db?cache=shared&mode=memory"))
	if err != nil {
		b.Fatal(err)
	}
	_, err = db.db.Exec(TestModel{}.CreateSQL())
	if err != nil {
		b.Fatal(err)
	}

	res, err := db.db.Exec("INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`)"+
		"VALUES (?,?,?,?)", 12, "Deng", 18, "Ming")

	if err != nil {
		b.Fatal(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		b.Fatal(err)
	}
	if affected == 0 {
		b.Fatal()
	}

	b.Run("unsafe", func(b *testing.B) {
		db.creator = valuer.NewUnsafeValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("reflect", func(b *testing.B) {
		db.creator = valuer.NewReflectValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func (TestModel) CreateSQL() string {
	return `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL
)
`
}
