package main

import (
	"fmt"
	"log"
	"slices"
	"strings"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

const (
	notAssigned = ""
)

func getTasks(sender string) string {
	tasksInfo.RLock()
	tasksIDs := make([]int, 0, len(tasksInfo.TasksNames))
	for taskID := range tasksInfo.TasksNames {
		tasksIDs = append(tasksIDs, taskID)
	}
	tasksInfo.RUnlock()

	if len(tasksIDs) == 0 {
		return "Нет задач"
	}

	slices.Sort(tasksIDs)

	var result strings.Builder
	for _, ID := range tasksIDs {
		tasksInfo.RLock()
		taskName := tasksInfo.TasksNames[ID]
		taskCreator := tasksInfo.TasksCreators[ID]
		taskExecutor := tasksInfo.TasksAssignee[ID]
		tasksInfo.RUnlock()

		result.WriteString(fmt.Sprintf("%d. %s by %s\n", ID, taskName, taskCreator))

		switch taskExecutor {
		case notAssigned:
			result.WriteString(fmt.Sprintf("/assign_%d", ID))
		case sender:
			result.WriteString(fmt.Sprintf("assignee: я\n/unassign_%d /resolve_%d", ID, ID))
		default:
			result.WriteString(fmt.Sprintf("assignee: %s", taskExecutor))
		}

		result.WriteString("\n\n")
	}

	out := result.String()
	return strings.TrimSuffix(out, "\n\n")
}

func createTask(taskName string, sender string) string {
	var ID int

	taskCounter.Lock()
	taskCounter.TasksCounter += 1
	ID = taskCounter.TasksCounter
	taskCounter.Unlock()

	tasksInfo.Lock()
	tasksInfo.TasksNames[ID] = taskName
	tasksInfo.TasksCreators[ID] = sender
	tasksInfo.TasksAssignee[ID] = ""
	tasksInfo.Unlock()

	result := fmt.Sprintf("Задача \"%s\" создана, id=%d", taskName, ID)

	return result
}

func assignTask(id int, sender string, bot *tgbotapi.BotAPI) string {
	var taskName string
	var ok bool

	tasksInfo.RLock()
	if taskName, ok = tasksInfo.TasksNames[id]; !ok {
		tasksInfo.RUnlock()
		return "Нет такой задачи"
	}
	taskExecutor := tasksInfo.TasksAssignee[id]
	tasksInfo.RUnlock()

	result := fmt.Sprintf("Задача \"%s\" назначена на вас", taskName)

	tasksInfo.Lock()
	tasksInfo.TasksAssignee[id] = sender
	tasksInfo.Unlock()

	if taskExecutor == "" {
		tasksInfo.RLock()
		taskCreator := tasksInfo.TasksCreators[id]
		tasksInfo.RUnlock()

		usersChats.RLock()
		creatorChatID := usersChats.UserChats[taskCreator]
		usersChats.RUnlock()

		messageToCreator := fmt.Sprintf("Задача \"%s\" назначена на %s", taskName, sender)
		_, err := bot.Send(tgbotapi.NewMessage(
			creatorChatID,
			messageToCreator,
		))
		if err != nil {
			log.Println(err)
		}
	} else if taskExecutor != sender {
		usersChats.RLock()
		executorChatID := usersChats.UserChats[taskExecutor]
		usersChats.RUnlock()

		messageToExecutor := fmt.Sprintf("Задача \"%s\" назначена на %s", taskName, sender)

		_, err := bot.Send(tgbotapi.NewMessage(
			executorChatID,
			messageToExecutor,
		))
		if err != nil {
			log.Println(err)
		}
	}

	return result
}

func unassignTask(id int, sender string, bot *tgbotapi.BotAPI) string {
	tasksInfo.Lock()
	if executor, ok := tasksInfo.TasksAssignee[id]; !ok || executor != sender {
		tasksInfo.Unlock()
		return "Задача не на вас"
	}
	tasksInfo.TasksAssignee[id] = ""
	tasksInfo.Unlock()

	tasksInfo.RLock()
	taskName := tasksInfo.TasksNames[id]
	creator := tasksInfo.TasksCreators[id]
	tasksInfo.RUnlock()

	usersChats.RLock()
	creatorChatID := usersChats.UserChats[creator]
	usersChats.RUnlock()

	messageToCreator := fmt.Sprintf("Задача \"%s\" осталась без исполнителя", taskName)
	_, err := bot.Send(tgbotapi.NewMessage(
		creatorChatID,
		messageToCreator,
	))
	if err != nil {
		log.Println(err)
	}

	return "Принято"
}

func resolveTask(id int, sender string, bot *tgbotapi.BotAPI) string {
	tasksInfo.Lock()
	if executor, ok := tasksInfo.TasksAssignee[id]; !ok || executor != sender {
		tasksInfo.Unlock()
		return "Задача не на вас"
	}
	delete(tasksInfo.TasksAssignee, id)
	taskName := tasksInfo.TasksNames[id]
	creator := tasksInfo.TasksCreators[id]
	delete(tasksInfo.TasksNames, id)
	delete(tasksInfo.TasksCreators, id)
	tasksInfo.Unlock()

	usersChats.RLock()
	creatorChatID := usersChats.UserChats[creator]
	usersChats.RUnlock()

	messageToCreator := fmt.Sprintf("Задача \"%s\" выполнена %s", taskName, sender)
	_, err := bot.Send(tgbotapi.NewMessage(
		creatorChatID,
		messageToCreator,
	))
	if err != nil {
		log.Println(err)
	}

	return fmt.Sprintf("Задача \"%s\" выполнена", taskName)
}

func getMyTasks(sender string) string {
	var tasksIDs []int

	tasksInfo.RLock()
	for taskID, taskExecutor := range tasksInfo.TasksAssignee {
		if taskExecutor == sender {
			tasksIDs = append(tasksIDs, taskID)
		}
	}
	tasksInfo.RUnlock()

	slices.Sort(tasksIDs)

	var result strings.Builder

	tasksInfo.RLock()
	for _, ID := range tasksIDs {
		taskName := tasksInfo.TasksNames[ID]
		result.WriteString(fmt.Sprintf("%d. %s by %s\n/unassign_%d /resolve_%d\n", ID, taskName, sender, ID, ID))
	}
	tasksInfo.RUnlock()

	out := result.String()
	return strings.TrimSuffix(out, "\n")
}

func getOwnTasks(sender string) string {
	var tasksIDs []int

	tasksInfo.RLock()
	for taskID, taskCreator := range tasksInfo.TasksCreators {
		if taskCreator == sender {
			tasksIDs = append(tasksIDs, taskID)
		}
	}
	tasksInfo.RUnlock()

	slices.Sort(tasksIDs)

	var result strings.Builder

	tasksInfo.RLock()
	for _, ID := range tasksIDs {
		taskName := tasksInfo.TasksNames[ID]
		result.WriteString(fmt.Sprintf("%d. %s by %s\n/assign_%d\n", ID, taskName, sender, ID))
	}
	tasksInfo.RUnlock()

	out := result.String()
	return strings.TrimSuffix(out, "\n")
}
