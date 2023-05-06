package reflect

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIterateArrayOrSlice(t *testing.T) {
	testCase := []struct {
		name      string
		entity    any
		wantVales []any
		wantErr   error
	}{
		{
			name:      "Array",
			entity:    [3]int{1, 2, 3},
			wantVales: []any{1, 2, 3},
		},
		{
			name:      "Slice",
			entity:    []int{1, 2, 3},
			wantVales: []any{1, 2, 3},
		},
		{
			name:    "int",
			entity:  15,
			wantErr: errors.New("不是数组或切片类型"),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			vals, err := IterateArrayOrSlice(tc.entity)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, vals, tc.wantVales)
		})
	}
}

func TestIterateMap(t *testing.T) {
	testCase := []struct {
		name       string
		entity     any
		wantKeys   []string
		wantValues []any
		wantErr    error
	}{
		{name: "map",
			entity: map[string]string{
				"A": "a",
				"B": "b",
			},
			wantKeys:   []string{"A", "B"},
			wantValues: []any{"a", "b"}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			keys, values, err := IterateMap(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.EqualValues(t, keys, tc.wantKeys)
			assert.EqualValues(t, values, tc.wantValues)
		})
	}
}
