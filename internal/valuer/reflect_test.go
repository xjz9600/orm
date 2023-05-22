package valuer

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"orm/model"
	"testing"
)

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  sql.NullString
}

func Test_reflectValue_SetColumns(t *testing.T) {
	SetColumns(t, NewReflectValue)
}

func SetColumns(t *testing.T, creator Creator) {
	testCase := []struct {
		name       string
		entity     any
		rows       *sqlmock.Rows
		wantErr    error
		wantEntity any
	}{
		{
			name:   "set columns",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				rows.AddRow(1, "Tom", 18, sql.NullString{Valid: true, String: "Jerry"})
				return rows
			}(),
			wantEntity: &TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  sql.NullString{Valid: true, String: "Jerry"},
			},
		},
		{
			name:   "partial columns",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "last_name"})
				rows.AddRow(1, sql.NullString{Valid: true, String: "Jerry"})
				return rows
			}(),
			wantEntity: &TestModel{
				Id:       1,
				LastName: sql.NullString{Valid: true, String: "Jerry"},
			},
		},
	}
	r := model.NewRegistry()
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			model, err := r.Get(tc.entity)
			require.NoError(t, err)
			val := creator(model, tc.entity)
			mockRows := tc.rows
			mock.ExpectQuery("SELECT XXX").WillReturnRows(mockRows)
			rows, err := mockDB.Query("SELECT XXX")
			require.NoError(t, err)
			rows.Next()
			err = val.SetColumns(rows)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, tc.entity, tc.wantEntity)
		})
	}
}
