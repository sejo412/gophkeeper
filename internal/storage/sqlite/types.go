package sqlite

import (
	"github.com/sejo412/gophkeeper/internal/models"
)

type table int

const (
	tableUnknown table = iota
	tableUsers
	tablePasswords
	tableTexts
	tableBins
	tableBanks
)

const (
	tableUnknownName   string = "unknown"
	tableUsersName     string = "users"
	tablePasswordsName string = "passwords"
	tableTextsName     string = "texts"
	tableBinsName      string = "bins"
	tableBanksName     string = "banks"
)

type query struct {
	table table
	query string
	args  []any
}

func (t table) String() string {
	switch t {
	case tableUnknown:
		return tableUnknownName
	case tableUsers:
		return tableUsersName
	case tablePasswords:
		return tablePasswordsName
	case tableTexts:
		return tableTextsName
	case tableBins:
		return tableBinsName
	case tableBanks:
		return tableBanksName
	default:
		return tableUnknownName
	}
}

func tables(r models.RecordType) table {
	switch r {
	case models.RecordPassword:
		return tablePasswords
	case models.RecordText:
		return tableTexts
	case models.RecordBin:
		return tableBins
	case models.RecordBank:
		return tableBanks
	default:
		return tableUnknown
	}
}
