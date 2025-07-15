package models

type User struct {
	ID          int
	Email       string
	PassHash    []byte
	Name        string
	Surname     string
	MiddleName  string
	PhoneNumber string
	BirthDate   BirthDate
}

type Student struct {
	User
	Group         string
	StudentNumber string
}

type Teacher struct {
	User
	Title      string
	Department string
	Degree     string
}

type Admin struct {
	Email    string
	PassHash string
}
