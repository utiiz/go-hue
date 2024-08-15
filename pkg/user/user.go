package user

type User struct {
	Username string
}

func NewUser(username string) *User {
	return &User{
		Username: username,
	}
}

func (u *User) String() string {
	return u.Username
}
