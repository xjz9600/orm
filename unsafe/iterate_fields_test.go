package unsafe

import "testing"

func TestPrintFieldOffset(t *testing.T) {
	testCase := []struct {
		name   string
		entity any
	}{
		{
			name:   "User",
			entity: User{},
		},
		{
			name:   "UserV1",
			entity: UserV1{},
		}, // TODO: Add test cases.
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			PrintFieldOffset(tc.entity)
		})
	}
}

type User struct {
	Name    string
	Age     int32
	Alias   []string
	Address string
}

type UserV1 struct {
	Name    string
	Age     int32
	AgeV1   int32
	Alias   []string
	Address string
}
