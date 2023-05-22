package sql

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSQLMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	defer db.Close()
	require.NoError(t, err)
	mockRows := mock.NewRows([]string{"id", "first_name"})
	mockRows.AddRow(1, "Tom")
	mockRows.AddRow(2, "Jerry")
	mock.ExpectQuery("SELECT id,first_name FROM `user`.*").WithArgs(3).WillReturnRows(mockRows)
	rows, err := db.QueryContext(context.Background(), "SELECT id,first_name FROM `user` where id = ?", 3)
	require.NoError(t, err)
	for rows.Next() {
		tm := TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName)
		require.NoError(t, err)
	}
	var testErr = errors.New("mock error")
	mock.ExpectQuery("SELECT id FROM `user`.*").WillReturnError(testErr)
	row := db.QueryRowContext(context.Background(), "SELECT id FROM `user` where id = 3")
	assert.Equal(t, row.Err(), testErr)
}
