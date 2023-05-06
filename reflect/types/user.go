package types

import "fmt"

type User struct {
	Name string
	age  int
}

func NewUser(name string, age int) User {
	return User{
		name,
		age,
	}
}

func NewUserPtr(name string, age int) *User {
	return &User{
		name,
		age,
	}
}

func (u User) GetAge() int {
	return u.age
}
func (u *User) ChangeName(newName string) {
	u.Name = newName
}

func (u User) private() {
	fmt.Println("private")
}
