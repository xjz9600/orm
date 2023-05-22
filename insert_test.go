package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"orm/internal/errs"
	"testing"
)

func TestInserter_SQLite_upsert(t *testing.T) {
	db := memoryDB(t, DBWithDialect(sqliteDialect{}))
	testCase := []struct {
		name      string
		i         *Inserter[TestModel]
		wantErr   error
		wantQuery *Query
	}{
		{
			name: "upsert",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  sql.NullString{String: "Jerry", Valid: true},
			}).OnDuplicateKey().ConflictColum("Id").Update(Assign("FirstName", "Tim"), Assign("Age", 19)),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) " +
					"ON CONFLICT(`id`) DO UPDATE SET `first_name`=?,`age`=?;",
				Args: []any{int64(12), "Tom", int8(18), sql.NullString{String: "Jerry", Valid: true}, "Tim", 19},
			},
		},
		{
			name: "upsert values",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  sql.NullString{String: "Jerry", Valid: true},
			}).OnDuplicateKey().ConflictColum("Id").Update(C("FirstName"), C("Age")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) " +
					"ON CONFLICT(`id`) DO UPDATE SET `first_name`=exclude.`first_name`,`age`=exclude.`age`;",
				Args: []any{int64(12), "Tom", int8(18), sql.NullString{String: "Jerry", Valid: true}},
			},
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.i.Build()
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, q, tc.wantQuery)
		})
	}
}

func TestInserter_Build(t *testing.T) {
	db := memoryDB(t)
	testCase := []struct {
		name      string
		i         *Inserter[TestModel]
		wantErr   error
		wantQuery *Query
	}{
		{
			name:    "no row",
			i:       NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			name: "single row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  sql.NullString{String: "Jerry", Valid: true},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);",
				Args: []any{int64(12), "Tom", int8(18), sql.NullString{String: "Jerry", Valid: true}},
			},
		},
		{
			name: "multiple row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id:        13,
				FirstName: "Xiejunze",
				Age:       19,
				LastName:  sql.NullString{String: "Archer", Valid: true},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?);",
				Args: []any{int64(12), "Tom", int8(18), sql.NullString{String: "Jerry", Valid: true}, int64(13), "Xiejunze", int8(19), sql.NullString{String: "Archer", Valid: true}},
			},
		},
		{
			name: "partial row",
			i: NewInserter[TestModel](db).Columns("Id", "FirstName").Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id:        13,
				FirstName: "Xiejunze",
				Age:       19,
				LastName:  sql.NullString{String: "Archer", Valid: true},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`) VALUES (?,?),(?,?);",
				Args: []any{int64(12), "Tom", int64(13), "Xiejunze"},
			},
		},
		{
			name: "upsert",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  sql.NullString{String: "Jerry", Valid: true},
			}).OnDuplicateKey().Update(Assign("FirstName", "Tim"), Assign("Age", 19)),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`=?,`age`=?;",
				Args: []any{int64(12), "Tom", int8(18), sql.NullString{String: "Jerry", Valid: true}, "Tim", 19},
			},
		},
		{
			name: "upsert values",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  sql.NullString{String: "Jerry", Valid: true},
			}).OnDuplicateKey().Update(C("FirstName"), C("Age")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`age`=VALUES(`age`);",
				Args: []any{int64(12), "Tom", int8(18), sql.NullString{String: "Jerry", Valid: true}},
			},
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.i.Build()
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, q, tc.wantQuery)
		})
	}
}

func TestInserter_Exec(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	defer mockDb.Close()
	require.NoError(t, err)
	db, err := OpenDB(mockDb)
	require.NoError(t, err)
	testCases := []struct {
		name     string
		i        *Inserter[TestModel]
		wantErr  error
		affected int64
	}{
		{
			name: "db error",
			i: func() *Inserter[TestModel] {
				mock.ExpectExec("INSERT INTO .*").WithArgs(12, "Tom", 18, sql.NullString{String: "Jerry", Valid: true}).WillReturnError(errors.New("db error"))
				return NewInserter[TestModel](db).Values(&TestModel{
					Id:        12,
					FirstName: "Tom",
					Age:       18,
					LastName:  sql.NullString{String: "Jerry", Valid: true},
				})
			}(),
			wantErr: errors.New("db error"),
		},
		{
			name: "db error",
			i: func() *Inserter[TestModel] {
				res := driver.RowsAffected(1)
				mock.ExpectExec("INSERT INTO .*").WithArgs(12, "Tom", 18, sql.NullString{String: "Jerry", Valid: true}).WillReturnResult(res)
				return NewInserter[TestModel](db).Values(&TestModel{
					Id:        12,
					FirstName: "Tom",
					Age:       18,
					LastName:  sql.NullString{String: "Jerry", Valid: true},
				})
			}(),
			affected: 1,
		},
		{
			name: "query error",
			i: func() *Inserter[TestModel] {
				return NewInserter[TestModel](db).Values(&TestModel{}).Columns("Invalid")
			}(),
			wantErr: errs.NewErrUnKnownField("Invalid"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.i.Exec(context.Background())
			affected, err := res.RowsAffected()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				assert.Equal(t, affected, tc.affected)
			}
		})
	}
}
