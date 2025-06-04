package client

import (
	"bufio"
	"fmt"
	"os"

	"github.com/sejo412/gophkeeper/internal/models"
)

func mainMenu() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		clearScreen()
		fmt.Println(MainTitle.String())
		for i := MainTitle + 1; i <= MainExit; i++ {
			fmt.Printf("%d. %s\n", i, i.String())
		}
		fmt.Print("Choose: ")

		scanner.Scan()
		input := scanner.Text()

		switch input {
		case MainList.Key():
			listAllRecords()
		case MainPasswords.Key():
			subMenu(MainPasswords)
		case MainBanks.Key():
			subMenu(MainBanks)
		case MainTexts.Key():
			subMenu(MainTexts)
		case MainBins.Key():
			subMenu(MainBins)
		case MainExit.Key():
			fmt.Println("\nExiting...")
			os.Exit(0)
		default:
			fmt.Println("Unknown command. Try again.")
			waitForEnter()
		}
	}
}

func subMenu(parent MainMenu) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		clearScreen()
		fmt.Printf("%s %s\n", parent.String(), SubTitle.String())
		for i := SubTitle + 1; i <= SubExit; i++ {
			fmt.Printf("%d. %s\n", i, i.String())
		}
		fmt.Print("Choose: ")

		scanner.Scan()
		input := scanner.Text()

		switch input {
		case SubList.Key():
			stubFunction(parent.Record(), SubList.Action())
		case SubCreate.Key():
			stubFunction(parent.Record(), SubCreate.Action())
		case SubRead.Key():
			stubFunction(parent.Record(), SubRead.Action())
		case SubUpdate.Key():
			stubFunction(parent.Record(), SubUpdate.Action())
		case SubDelete.Key():
			stubFunction(parent.Record(), SubDelete.Action())
		case SubBack.Key():
			return
		case SubExit.Key():
			fmt.Println("\nExiting...")
			os.Exit(0)
		default:
			fmt.Println("Unknown command. Try again.")

		}
	}

}

func listAllRecords() {
	clearScreen()
	fmt.Println("\nListing all records (not implemented).")
	waitForEnter()
}

func stubFunction(object models.RecordType, action Action) {
	clearScreen()
	fmt.Printf("%s: %s\n", object.String(), action.String())
	fmt.Println("(not implemented)")
	waitForEnter()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func waitForEnter() {
	fmt.Print("\nPress Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
