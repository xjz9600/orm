package valuer

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"orm/model"
	"testing"
)

func BenchmarkSetColumns(b *testing.B) {
	fn := func(b *testing.B, creator Creator) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(b, err)
		defer mockDB.Close()
		mockRows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
		row := []driver.Value{1, "Tom", 18, "Jerry"}
		for i := 0; i < b.N; i++ {
			mockRows.AddRow(row...)
		}
		mock.ExpectQuery("SELECT XXx").WillReturnRows(mockRows)
		rows, err := mockDB.Query("SELECT XXx")
		require.NoError(b, err)
		r := model.NewRegistry()
		model, err := r.Get(&TestModel{})
		require.NoError(b, err)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rows.Next()
			val := creator(model, &TestModel{})
			_ = val.SetColumns(rows)
		}
	}
	b.Run("reflect", func(b *testing.B) {
		fn(b, NewReflectValue)
	})
	b.Run("unsafe", func(b *testing.B) {
		fn(b, NewUnsafeValue)
	})
}
