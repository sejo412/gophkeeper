package models

// RecordType is a type of record (password, text, etc).
type RecordType int

// Records type specific.
const (
	RecordUnknown RecordType = iota
	RecordPassword
	RecordText
	RecordBin
	RecordBank
)

// Names of RecordTypes.
const (
	RecordUnknownName  string = "unknown type"
	RecordPasswordName string = "password"
	RecordTextName     string = "text"
	RecordBinName      string = "binary data"
	RecordBankName     string = "bank's card"
)

// Meta type for meta field.
type Meta string

// ID type for record's ID field.
type ID int

// UserID type for record's uid field.
type UserID int

// User type fo user record.
type User struct {
	ID UserID
	Cn string
}

// Encrypted type for encrypted field in storage.
type Encrypted []byte

// Record type for clear record, includes all RecordType.
type Record struct {
	Password Password
	Text     Text
	Bin      Bin
	Bank     Bank
}

// RecordEncrypted type for encrypted ([]byte) record, includes all RecordType.
type RecordEncrypted struct {
	Password PasswordEncrypted
	Text     TextEncrypted
	Bin      BinEncrypted
	Bank     BankEncrypted
}

// RecordsEncrypted type for mass encrypted ([]byte) records, includes all RecordType.
type RecordsEncrypted struct {
	Password []PasswordEncrypted
	Text     []TextEncrypted
	Bin      []BinEncrypted
	Bank     []BankEncrypted
}

// Password type for password field in Record.
type Password struct {
	ID       ID
	Login    string
	Password string
	Meta     Meta
}

// PasswordEncrypted type for password field in RecordEncrypted.
type PasswordEncrypted struct {
	ID       ID
	Login    Encrypted
	Password Encrypted
	Meta     Encrypted
}

// Text type for text field in Record.
type Text struct {
	ID   ID
	Text string
	Meta Meta
}

// TextEncrypted type for text field in RecordEncrypted.
type TextEncrypted struct {
	ID   ID
	Text Encrypted
	Meta Encrypted
}

// Bin type for bin field in Record.
type Bin struct {
	ID   ID
	Data []byte
	Meta Meta
}

// BinEncrypted type for bin field in RecordEncrypted.
type BinEncrypted struct {
	ID   ID
	Data Encrypted
	Meta Encrypted
}

// Bank type for bank field in Record.
type Bank struct {
	ID     ID
	Number string
	Name   string
	Date   string
	Cvv    string
	Meta   Meta
}

// BankEncrypted type for bank field in RecordEncrypted.
type BankEncrypted struct {
	ID     ID
	Number Encrypted
	Name   Encrypted
	Date   Encrypted
	Cvv    Encrypted
	Meta   Encrypted
}

// String implements Stringer interface.
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
