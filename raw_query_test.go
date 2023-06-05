package orm

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRawQuery_GET(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	defer mockDb.Close()
	require.NoError(t, err)
	db, err := OpenDB(mockDb)
	require.NoError(t, err)
	mock.ExpectQuery("SELECT .*").WillReturnError(errors.New("query error"))

	mock.ExpectQuery("SELECT .*").WithArgs(-1).WillReturnRows(mock.NewRows([]string{"id", "first_name", "age", "last_name"}))

	mockRows := mock.NewRows([]string{"id", "first_name", "age", "last_name"})
	mockRows.AddRow(1, "Tom", 18, "Jerry")
	mock.ExpectQuery("SELECT .*").WithArgs(1).WillReturnRows(mockRows)
	testCase := []struct {
		name    string
		r       *RawQuerier[TestModel]
		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "query err",
			r:       RawQuery[TestModel](db, "SELECT * FROM `test_model`"),
			wantErr: errors.New("query error"),
		},
		{
			name:    "no rows",
			r:       RawQuery[TestModel](db, "SELECT * FROM `test_model`WHERE `id` = ?", -1),
			wantErr: ErrNoRows,
		},
		{
			name: "data",
			r:    RawQuery[TestModel](db, "SELECT * FROM `test_model`WHERE `id` = ?", 1),
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
			model, err := tc.r.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, model)
		})
	}
}
