package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	initGame()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "exit" {
			break
		}
		fmt.Println(handleCommand(line))
	}
}

func initGame() {
	var kitchen Room
	var corridor Room
	var street Room
	var appartment Room

	tea := Item{
		Name:         "чай",
		ItemPosition: "на столе",
	}

	key := Item{
		Name:         "ключи",
		ItemPosition: "на столе",
		CanApply: map[string]string{
			"дверь": "дверь открыта",
		},
		UseItem: func(user *User) {
			for key := range user.Position.ConnectionsSet {
				user.Position.ConnectionsSet[key] = true
			}
		},
	}

	univer := Item{
		Name:       "универ",
		ItemAction: "идти в",
	}

	notes := Item{
		Name:         "конспекты",
		ItemPosition: "на столе",
	}

	backpack := Item{
		Name:         "рюкзак",
		ItemAction:   "собрать",
		ItemPosition: "на стуле",
	}

	corridor = Room{
		Name:            "коридор",
		WalkDescription: "ничего интересного",
		Connections:     []*Room{&kitchen, &appartment, &street},
		ConnectionsSet: map[*Room]bool{
			&kitchen:    true,
			&appartment: true,
			&street:     false,
		},
	}

	street = Room{
		Name:            "улица",
		WalkDescription: "на улице весна",
		Connections:     []*Room{&corridor},
		ConnectionsSet: map[*Room]bool{
			&corridor: false,
		},
		IsOutsideHome: true,
	}

	appartment = Room{
		Name:            "комната",
		WalkDescription: "ты в своей комнате",
		Connections:     []*Room{&corridor},
		ConnectionsSet: map[*Room]bool{
			&corridor: true,
		},

		Items: []*RoomTarget{
			{
				Item:           &key,
				IsNeededToFind: false,
			},
			{
				Item:           &notes,
				IsNeededToFind: false,
			},
			{
				Item:           &backpack,
				IsNeededToFind: false,
			},
		},
	}

	kitchen = Room{
		Name:            "кухня",
		Description:     "ты находишься на кухне, ",
		WalkDescription: "кухня, ничего интересного",

		Items: []*RoomTarget{
			{
				Item:           &tea,
				IsNeededToFind: false,
			},
			{
				Item:           &backpack,
				IsNeededToFind: true,
			},
			{
				Item:           &univer,
				IsNeededToFind: true,
			},
		},

		Connections: []*Room{&corridor},
		ConnectionsSet: map[*Room]bool{
			&corridor: true,
		},

		IsOutsideHome: false,
	}

	user.SetPosition(&kitchen)
	user.Storage = &backpack
	user.ClearItems()

	gameMap.ClearMap()
	gameMap.AddRoom(&corridor)
	gameMap.AddRoom(&kitchen)
	gameMap.AddRoom(&appartment)
	gameMap.AddRoom(&street)

}

func handleCommand(command string) string {
	parameters := strings.Split(command, " ")
	cmd := parameters[0]
	args := parameters[1:]

	commandResult, err := user.DoCommand(cmd, args...)
	if err != nil {
		fmt.Println(err)
	}

	return commandResult
}
