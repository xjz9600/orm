package sql

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestDB(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL,
    json_data TEXT NOT NULL
)
`)
	require.NoError(t, err)
	res, err := db.ExecContext(ctx, "INSERT INTO test_model(`id`, `first_name`, `age`, `last_name`,`json_data`) VALUES(?, ?, ?, ?,?)", 1, "Tom", 18, "Jerry", []byte(`{"Name":"Tom"}`))
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	assert.EqualValues(t, 1, affected)
	lastId, err := res.LastInsertId()
	require.NoError(t, err)
	assert.EqualValues(t, 1, lastId)
	row := db.QueryRowContext(ctx, "SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?", 1)
	tm := TestModel{}
	err = row.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
	require.NoError(t, err)

	row = db.QueryRowContext(ctx, "SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?", 2)
	tm = TestModel{}
	err = row.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
	assert.Equal(t, err, sql.ErrNoRows)

	rows, err := db.QueryContext(ctx, "SELECT `id`, `first_name`, `age`, `last_name`,`json_data` FROM `test_model` WHERE `id` = ?", 1)
	require.NoError(t, err)
	for rows.Next() {
		tm = TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName, &tm.JsonData)
		assert.Equal(t, tm.JsonData.Valid, true)
		assert.Equal(t, tm.JsonData.Val.Name, "Tom")
		require.NoError(t, err)
	}
	cancel()
}

func TestTx(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL,
    json_data TEXT NOT NULL
)
`)
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	res, err := tx.ExecContext(ctx, "INSERT INTO test_model(`id`, `first_name`, `age`, `last_name`,`json_data`) VALUES(?, ?, ?, ?,?)", 1, "Tom", 18, "Jerry", []byte(`{"Name":"Tom"}`))
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			log.Println(err)
		}
	}
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	log.Println("受影响行数", affected)
	lastId, err := res.LastInsertId()
	require.NoError(t, err)
	log.Println("最后插入 ID", lastId)

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}
	cancel()
}
func TestPrepareStatement(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL,
    json_data TEXT NOT NULL
)
`)
	stmt, err := db.PrepareContext(ctx, "INSERT INTO test_model(`id`, `first_name`, `age`, `last_name`,`json_data`) VALUES(?, ?, ?, ?, ?)")
	require.NoError(t, err)
	_, err = stmt.ExecContext(ctx, 1, "Tom", 18, "Jerry", JsonColum[User]{Valid: true, Val: User{Name: "Tom"}})
	require.NoError(t, err)
	cancel()
	stmt.Close()
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  sql.NullString
	JsonData  JsonColum[User]
}

type User struct {
	Name string
}
