package client

import (
	"strconv"

	"github.com/sejo412/gophkeeper/internal/models"
	pb "github.com/sejo412/gophkeeper/proto"
)

type MainMenu int
type SubMenu int

type Action int

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

func protoID(value int) *int64 {
	res := new(int64)
	*res = int64(value)
	return res
}
