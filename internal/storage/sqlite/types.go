package sqlite

import (
	"github.com/sejo412/gophkeeper/internal/models"
)

type table int

const (
	tableUnknown table = iota
	tableUsers
	tablePasswords
	tablePasswordsMap
	tableTexts
	tableTextsMap
	tableBins
	tableBinsMap
	tableBanks
	tableBanksMap
)

const (
	tableUnknownName      string = "unknown"
	tableUsersName        string = "users"
	tablePasswordsName    string = "passwords"
	tablePasswordsMapName string = "passwords_map"
	tableTextsName        string = "texts"
	tableTextsMapName     string = "texts_map"
	tableBinsName         string = "bins"
	tableBinsMapName      string = "bins_map"
	tableBanksName        string = "banks"
	tableBanksMapName     string = "banks_map"
)

type query struct {
	table table
	query string
	args  []any
}

type recordTables struct {
	record    table
	recordMap table
}

func (t table) String() string {
	switch t {
	case tableUnknown:
		return tableUnknownName
	case tableUsers:
		return tableUsersName
	case tablePasswords:
		return tablePasswordsName
	case tablePasswordsMap:
		return tablePasswordsMapName
	case tableTexts:
		return tableTextsName
	case tableTextsMap:
		return tableTextsMapName
	case tableBins:
		return tableBinsName
	case tableBinsMap:
		return tableBinsMapName
	case tableBanks:
		return tableBanksName
	case tableBanksMap:
		return tableBanksMapName
	default:
		return tableUnknownName
	}
}

// Deprecated
func tableByType(t models.RecordType) table {
	switch t {
	case models.RecordPassword:
		return tablePasswords
	default:
		return tableUnknown
	}
}

func tables(r models.RecordType) recordTables {
	switch r {
	case models.RecordPassword:
		return recordTables{
			record:    tablePasswords,
			recordMap: tablePasswordsMap,
		}
	case models.RecordText:
		return recordTables{
			record:    tableTexts,
			recordMap: tableTextsMap,
		}
	case models.RecordBin:
		return recordTables{
			record:    tableBins,
			recordMap: tableBinsMap,
		}
	case models.RecordBank:
		return recordTables{
			record:    tableBanks,
			recordMap: tableBanksMap,
		}
	default:
		return recordTables{}
	}
}
