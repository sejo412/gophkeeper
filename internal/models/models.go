package models

type RecordType int

const (
	RecordUnknown RecordType = iota
	RecordPassword
	RecordText
	RecordBin
	RecordBank
)

const (
	RecordUnknownName  string = "unknown type"
	RecordPasswordName string = "password"
	RecordTextName     string = "text"
	RecordBinName      string = "binary data"
	RecordBankName     string = "bank's card"
)

type Meta string

type ID int

type UserID int

type Map struct {
	ID     ID
	UserID UserID
}
type User struct {
	ID UserID
	Cn string
}

type Encrypted []byte

type Record struct {
	Password Password
	Text     Text
	Bin      Bin
	Bank     Bank
}

type RecordEncrypted struct {
	Password PasswordEncrypted
	Text     TextEncrypted
	Bin      BinEncrypted
	Bank     BankEncrypted
}

type RecordsEncrypted struct {
	Password []PasswordEncrypted
	Text     []TextEncrypted
	Bin      []BinEncrypted
	Bank     []BankEncrypted
}

type Password struct {
	ID       ID
	Login    string
	Password string
	Meta     Meta
}

type PasswordEncrypted struct {
	ID       ID
	Login    Encrypted
	Password Encrypted
	Meta     Encrypted
}

type Text struct {
	ID   ID
	Text string
	Meta Meta
}

type TextEncrypted struct {
	ID   ID
	Text Encrypted
	Meta Encrypted
}

type Bin struct {
	ID   ID
	Data []byte
	Meta Meta
}

type BinEncrypted struct {
	ID   ID
	Data Encrypted
	Meta Encrypted
}

type Bank struct {
	ID     ID
	Number string
	Date   string
	Cvv    string
	Meta   Meta
}

type BankEncrypted struct {
	ID     ID
	Number Encrypted
	Date   Encrypted
	Cvv    Encrypted
	Meta   Encrypted
}

func (r RecordType) String() string {
	switch r {
	case RecordUnknown:
		return RecordUnknownName
	case RecordPassword:
		return RecordPasswordName
	case RecordText:
		return RecordTextName
	case RecordBin:
		return RecordBinName
	case RecordBank:
		return RecordBankName
	default:
		return RecordUnknownName
	}
}
