package sql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJsonColumn_Value(t *testing.T) {
	js := JsonColum[User]{Valid: true, Val: User{Name: "Tom"}}
	value, err := js.Value()
	assert.Nil(t, err)
	assert.Equal(t, []byte(`{"Name":"Tom"}`), value)
	js = JsonColum[User]{}
	value, err = js.Value()
	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestJsonColumn_Scan(t *testing.T) {
	testCases := []struct {
		name    string
		src     any
		wantErr error
		wantVal User
		valid   bool
	}{
		{
			name: "nil",
		},
		{
			name:    "string",
			src:     `{"Name":"Tom"}`,
			wantVal: User{Name: "Tom"},
			valid:   true,
		},
		{
			name:    "bytes",
			src:     []byte(`{"Name":"Tom"}`),
			wantVal: User{Name: "Tom"},
			valid:   true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			js := &JsonColum[User]{}
			err := js.Scan(tc.src)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, js.Val)
			assert.Equal(t, tc.valid, js.Valid)
		})
	}
}

func TestJsonColumn_ScanTypes(t *testing.T) {
	jsSlice := JsonColum[[]string]{}
	err := jsSlice.Scan(`["a", "b", "c"]`)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, jsSlice.Val)
	val, err := jsSlice.Value()
	assert.Nil(t, err)
	assert.Equal(t, []byte(`["a","b","c"]`), val)

	jsMap := JsonColum[map[string]string]{}
	err = jsMap.Scan(`{"a":"a value"}`)
	assert.Nil(t, err)
	val, err = jsMap.Value()
	assert.Nil(t, err)
	assert.Equal(t, []byte(`{"a":"a value"}`), val)
}

func TestExampleJsonColumn_Value(t *testing.T) {
	js := JsonColum[User]{Valid: true, Val: User{Name: "Tom"}}
	value, err := js.Value()
	assert.Nil(t, err)
	assert.Equal(t, string(value.([]byte)), `{"Name":"Tom"}`)
}

func TestExampleJsonColumn_Scan(t *testing.T) {
	js := JsonColum[User]{}
	err := js.Scan(`{"Name":"Tom"}`)
	assert.Nil(t, err)
	assert.Equal(t, js.Val.Name, "Tom")
}
