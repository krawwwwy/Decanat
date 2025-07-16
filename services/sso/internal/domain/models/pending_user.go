package models

type PendingUser struct {
	Email       string            `json:"email"`
	PassHash    []byte            `json:"pass_hash"`
	Name        string            `json:"name"`
	Surname     string            `json:"surname"`
	MiddleName  string            `json:"middle_name"`
	PhoneNumber string            `json:"phone_number"`
	BirthDate   BirthDate         `json:"birth_date"`
	Role        string            `json:"role"` // "student", "teacher", "admin"
	Meta        map[string]string `json:"meta"` // role-specific fields
}
