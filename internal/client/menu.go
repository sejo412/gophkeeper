package client

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/sejo412/gophkeeper/internal/models"
)

func mainMenu(ctx context.Context, c *Client) {
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
			subMenu(ctx, c, MainPasswords)
		case MainBanks.Key():
			subMenu(ctx, c, MainBanks)
		case MainTexts.Key():
			subMenu(ctx, c, MainTexts)
		case MainBins.Key():
			subMenu(ctx, c, MainBins)
		case MainExit.Key():
			fmt.Println("\nExiting...")
			os.Exit(0)
		default:
			fmt.Println("Unknown command. Try again.")
			waitForEnter()
		}
	}
}

func subMenu(ctx context.Context, c *Client, parent MainMenu) {
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
			actionFunction(ctx, c, parent.Record(), SubList.Action())
		case SubCreate.Key():
			actionFunction(ctx, c, parent.Record(), SubCreate.Action())
		case SubRead.Key():
			actionFunction(ctx, c, parent.Record(), SubRead.Action())
		case SubUpdate.Key():
			actionFunction(ctx, c, parent.Record(), SubUpdate.Action())
		case SubDelete.Key():
			actionFunction(ctx, c, parent.Record(), SubDelete.Action())
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

func actionFunction(ctx context.Context, c *Client, object models.RecordType, action Action) {
	clearScreen()
	fmt.Printf("%s: %s\n", object.String(), action.String())
	scanner := bufio.NewScanner(os.Stdin)
	switch object {
	case models.RecordPassword:
		switch action {
		case ActionCreate:
			createPassword(ctx, c, scanner)
		case ActionRead:
			getPassword(ctx, c, scanner)
		default:
			fmt.Printf("(action %q not supported for %q)\n", action.String(), object.String())
		}
	default:
		fmt.Println("(not implemented)")
	}
	waitForEnter()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func waitForEnter() {
	fmt.Print("\nPress Enter to continue...")
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}
