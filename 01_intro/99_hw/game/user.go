package main

import (
	"slices"
	"strings"
)

var user User

type User struct {
	Position *Room
	Storage  *Item
	Items    map[*Item]struct{}
}

func (user *User) SetPosition(room *Room) {
	user.Position = room
}

func (user *User) ItemRoomIndex(item string) (int, error) {
	itemIndex := slices.IndexFunc(user.Position.Items, func(n *RoomTarget) bool {
		return n.Item.Name == item
	})

	if itemIndex == -1 {
		return itemIndex, errItemNotInRoom
	}

	return itemIndex, nil
}

func (user *User) TakeItem(item string) error {
	itemIndex, err := user.ItemRoomIndex(item)
	if err != nil {
		return err
	}

	user.Items[user.Position.Items[itemIndex].Item] = struct{}{}
	user.Position.Items[itemIndex].IsNeededToFind = false
	user.Position.Items = append(user.Position.Items[:itemIndex], user.Position.Items[itemIndex+1:]...)

	return nil
}

func (user *User) ClearItems() {
	user.Items = make(map[*Item]struct{})
}

func (user *User) DoCommand(command string, parameters ...string) (string, error) {
	var result string

	switch command {

	case CommandLookAround.CommandName:
		result = user.HandleLookAround(parameters...)

	case CommandWalk.CommandName:
		if len(parameters) < CommandWalk.ArgsNum {
			return "", errInvalidCommand
		}
		result = user.HandleWalk(parameters[0], parameters[1:]...)

	case CommandTake.CommandName, CommandWear.CommandName:
		if len(parameters) < CommandTake.ArgsNum {
			return "", errInvalidCommand
		}
		result = user.HandleTakeItem(command, parameters[0], parameters[1:]...)

	case CommandApply.CommandName:
		if len(parameters) < CommandApply.ArgsNum {
			return "", errInvalidCommand
		}
		result = user.HandleApply(parameters[0], parameters[1], parameters[2:]...)

	default:
		result = UnknownCommand.CommandName

	}

	return result, nil
}

func (user *User) HandleApply(what string, toWhat string, _ ...string) string {
	var result strings.Builder
	var item *Item
	for key := range user.Items {
		if key.Name == what {
			item = key
			break
		}
	}
	if item == nil {
		return "нет предмета в инвентаре - " + what
	}

	if item.CanApply != nil {
		if actionResult, ok := item.CanApply[toWhat]; !ok {
			result.WriteString("не к чему применить")
		} else {
			item.UseItem(user)
			result.WriteString(actionResult)
		}
	}

	return result.String()
}

func (user *User) HandleTakeItem(command string, what string, _ ...string) string {
	storeName := user.Storage.Name

	var checkStore bool
	for key := range user.Items {
		if key.Name == storeName {
			checkStore = true
			break
		}
	}

	if !checkStore && what != storeName {
		_, err := user.ItemRoomIndex(what)
		if err != nil {
			return err.Error()
		}

		return "некуда класть"
	}

	err := user.TakeItem(what)
	if err != nil {
		return err.Error()
	}

	var result strings.Builder

	switch command {
	case CommandWear.CommandName:
		result.WriteString("вы надели")

	case CommandTake.CommandName:
		result.WriteString("предмет добавлен в инвентарь")
	}

	result.WriteString(": " + what)

	return result.String()
}

func (user *User) HandleLookAround(_ ...string) string {
	var result strings.Builder

	if len(user.Position.Items) > 0 {
		result.WriteString(user.Position.Description)
	} else {
		result.WriteString("пустая комната")
	}

	result.WriteString(DisplayItems(user))
	result.WriteString(DisplayRoomTarget(user))
	result.WriteString(DisplayPossibleMoves(user))

	return result.String()
}

func (user *User) HandleWalk(where string, _ ...string) string {
	isDoorOpen, exists := user.Position.ConnectionsSet[gameMap.Rooms[where]]
	if !exists {
		return "нет пути в " + where
	}
	if !isDoorOpen {
		return "дверь закрыта"
	}

	user.SetPosition(gameMap.Rooms[where])

	var result strings.Builder

	result.WriteString(user.Position.WalkDescription)
	result.WriteString(DisplayPossibleMoves(user))

	return result.String()
}

func DisplayPossibleMoves(user *User) string {
	var builder strings.Builder

	builder.WriteString(". можно пройти - ")

	for i, room := range user.Position.Connections {
		if user.Position.IsOutsideHome {
			builder.WriteString("домой")
		} else {
			builder.WriteString(room.Name)
		}
		if i < len(user.Position.Connections)-1 {
			builder.WriteString(", ")
		}
	}

	return builder.String()
}

func DisplayItems(user *User) string {
	var builder strings.Builder
	var items []*Item

	for _, roomTarget := range user.Position.Items {
		if !roomTarget.IsNeededToFind {
			items = append(items, roomTarget.Item)
		}
	}

	if len(items) == 0 {
		return ""
	}

	itemPlace := items[0].ItemPosition
	builder.WriteString(items[0].ItemPosition + ": ")

	for i, item := range items {
		if item.ItemPosition != itemPlace {
			builder.WriteString(item.ItemPosition + ": ")
		}

		builder.WriteString(item.Name)

		if i < len(items)-1 {
			builder.WriteString(", ")
		}
	}

	return builder.String()
}

func DisplayRoomTarget(user *User) string {
	var builder strings.Builder
	var targetItems []*RoomTarget

	for _, roomTarget := range user.Position.Items {
		if roomTarget.IsNeededToFind {
			targetItems = append(targetItems, roomTarget)
		}
	}

	if len(targetItems) == 0 {
		return ""
	}

	builder.WriteString(", надо ")
	for i, targetItem := range targetItems {
		if _, have := user.Items[targetItem.Item]; have {
			continue
		}

		builder.WriteString(targetItem.Item.ItemAction + " " + targetItem.Item.Name)

		if i < len(targetItems)-1 {
			builder.WriteString(" и ")
		}
	}

	return builder.String()
}
