package models

type Meta string

type ID int

type UserID int

type Map struct {
	ID     ID
	UserID UserID
}
type User struct {
	ID ID
	Cn string
}

type Password struct {
	ID       ID
	Login    string
	Password string
	Meta     Meta
}

type Text struct {
	ID   ID
	Text string
	Meta Meta
}

type Bin struct {
	ID   ID
	Data []byte
	Meta Meta
}

type Bank struct {
	ID     ID
	Number string
	Date   string
	Cvv    string
	Meta   Meta
}
