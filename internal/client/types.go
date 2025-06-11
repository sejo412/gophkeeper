package client

import (
	"strconv"

	"github.com/sejo412/gophkeeper/internal/models"
	pb "github.com/sejo412/gophkeeper/proto"
)

type MainMenu int
type SubMenu int

type Action int

type Field int

const (
	MainTitle MainMenu = iota
	MainList
	MainPasswords
	MainBanks
	MainTexts
	MainBins
	MainExit
)

const (
	SubTitle SubMenu = iota
	SubList
	SubCreate
	SubRead
	SubUpdate
	SubDelete
	SubBack
	SubExit
)

const (
	MainTitleName     string = "Main menu:"
	MainListName      string = "List all records"
	MainPasswordsName string = "Passwords"
	MainBanksName     string = "Bank's Cards"
	MainTextsName     string = "Texts"
	MainBinsName      string = "Binary Data"
	MainExitName      string = "Exit"
)

const (
	SubTitleName  string = "Sub menu:"
	SubListName   string = "List"
	SubCreateName string = "Create"
	SubReadName   string = "Read"
	SubUpdateName string = "Update"
	SubDeleteName string = "Delete"
	SubBackName   string = "Back"
	SubExitName   string = "Exit"
)

const (
	ActionUnknown Action = iota
	ActionList
	ActionCreate
	ActionRead
	ActionUpdate
	ActionDelete
)

const (
	ActionListName   string = "List"
	ActionCreateName string = "Create"
	ActionReadName   string = "Read"
	ActionUpdateName string = "Update"
	ActionDeleteName string = "Delete"
)

const (
	FieldID Field = iota
	FieldLogin
	FieldPassword
	FieldText
	FieldData
	FieldNumber
	FieldName
	FieldDate
	FieldCVV
	FieldMeta
)

const (
	FieldIDName       string = "ID"
	FieldLoginName    string = "Login"
	FieldPasswordName string = "Password"
	FieldTextName     string = "Text"
	FieldDataName     string = "Data"
	FieldNumberName   string = "Number"
	FieldNameName     string = "Owner"
	FieldDateName     string = "Date"
	FieldCVVName      string = "CVV"
	FieldMetaName     string = "Meta"
)

func (f Field) String() string {
	switch f {
	case FieldID:
		return FieldIDName
	case FieldLogin:
		return FieldLoginName
	case FieldPassword:
		return FieldPasswordName
	case FieldText:
		return FieldTextName
	case FieldData:
		return FieldDataName
	case FieldNumber:
		return FieldNumberName
	case FieldName:
		return FieldNameName
	case FieldDate:
		return FieldDateName
	case FieldCVV:
		return FieldCVVName
	case FieldMeta:
		return FieldMetaName
	default:
		return "unknown field type"
	}
}

func (m MainMenu) String() string {
	switch m {
	case MainTitle:
		return MainTitleName
	case MainList:
		return MainListName
	case MainPasswords:
		return MainPasswordsName
	case MainBanks:
		return MainBanksName
	case MainTexts:
		return MainTextsName
	case MainBins:
		return MainBinsName
	case MainExit:
		return MainExitName
	default:
		return "Unknown menu"
	}
}

func (m MainMenu) Key() string {
	return strconv.Itoa(int(m))
}

func (m MainMenu) Record() models.RecordType {
	switch m {
	case MainPasswords:
		return models.RecordPassword
	case MainBanks:
		return models.RecordBank
	case MainTexts:
		return models.RecordText
	case MainBins:
		return models.RecordBin
	default:
		return models.RecordUnknown
	}
}

func (s SubMenu) String() string {
	switch s {
	case SubTitle:
		return SubTitleName
	case SubList:
		return SubListName
	case SubCreate:
		return SubCreateName
	case SubRead:
		return SubReadName
	case SubUpdate:
		return SubUpdateName
	case SubDelete:
		return SubDeleteName
	case SubBack:
		return SubBackName
	case SubExit:
		return SubExitName
	default:
		return "Unknown sub menu"
	}
}

func (s SubMenu) Action() Action {
	switch s {
	case SubList:
		return ActionList
	case SubCreate:
		return ActionCreate
	case SubRead:
		return ActionRead
	case SubUpdate:
		return ActionUpdate
	case SubDelete:
		return ActionDelete
	default:
		return ActionUnknown
	}
}

func (s SubMenu) Key() string {
	return strconv.Itoa(int(s))
}

func (a Action) String() string {
	switch a {
	case ActionList:
		return ActionListName
	case ActionCreate:
		return ActionCreateName
	case ActionRead:
		return ActionReadName
	case ActionUpdate:
		return ActionUpdateName
	case ActionDelete:
		return ActionDeleteName
	default:
		return "Unknown action"
	}
}

func protoRecordType(value pb.RecordType) *pb.RecordType {
	return &value
}

func modelRecordTypeToProto(value models.RecordType) pb.RecordType {
	switch value {
	case models.RecordPassword:
		return pb.RecordType_PASSWORD
	case models.RecordText:
		return pb.RecordType_TEXT
	case models.RecordBin:
		return pb.RecordType_BIN
	case models.RecordBank:
		return pb.RecordType_BANK
	default:
		return pb.RecordType_UNKNOWN
	}
}

func protoID(value int) *int64 {
	res := new(int64)
	*res = int64(value)
	return res
}

func fields(t models.RecordType) []Field {
	switch t {
	case models.RecordPassword:
		return []Field{
			FieldLogin,
			FieldPassword,
			FieldMeta,
		}
	case models.RecordText:
		return []Field{
			FieldText,
			FieldMeta,
		}
	case models.RecordBin:
		return []Field{
			FieldData,
			FieldMeta,
		}
	case models.RecordBank:
		return []Field{
			FieldNumber,
			FieldName,
			FieldDate,
			FieldCVV,
			FieldMeta,
		}
	default:
		return nil
	}
}
