package main

import (
	"fmt"
	"log"
	"slices"
	"sync"
)

func RunPipeline(cmds ...cmd) {
	var in chan interface{}
	var out chan interface{}

	wg := &sync.WaitGroup{}
	wg.Add(len(cmds))

	for _, command := range cmds {
		out = make(chan interface{})

		go func(in, out chan interface{}, c cmd) {
			defer wg.Done()
			defer close(out)
			c(in, out)
		}(in, out, command)

		in = out
	}

	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	// 	in - string
	// 	out - User
	checkAlias := map[uint64]struct{}{}

	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for userEmail := range in {

		userEmailStr, ok := userEmail.(string)
		if !ok {
			log.Println("Не удалось преобразовать email пользователя в строку")
			continue
		}

		wg.Add(1)
		go func(email string) {
			defer wg.Done()

			user := GetUser(email)

			var canSend bool

			mu.Lock()
			if _, exist := checkAlias[user.ID]; !exist {
				checkAlias[user.ID] = struct{}{}
				canSend = true
			}
			mu.Unlock()

			if canSend {
				out <- user
			}

		}(userEmailStr)
	}

	wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User
	// 	out - MsgID
	users := make([]User, 0, GetMessagesMaxUsersBatch)

	wg := &sync.WaitGroup{}

	getMessagesWorker := func(batch ...User) {
		defer wg.Done()

		msgIDs, err := GetMessages(batch...)
		if err != nil {
			log.Println("Ошибка при получении сообщений")
			return
		}
		for _, msgID := range msgIDs {
			out <- msgID
		}
	}

	for user := range in {
		userFrame, ok := user.(User)
		if !ok {
			log.Println("В канал пришли данные, которые нельзя преобразовать к структуре пользователя")
			continue
		}

		users = append(users, userFrame)

		if len(users) < GetMessagesMaxUsersBatch {
			continue
		}

		wg.Add(1)
		go getMessagesWorker(users...)

		users = []User{}
	}

	if len(users) > 0 {
		wg.Add(1)
		go getMessagesWorker(users...)
	}

	wg.Wait()
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
	wg := &sync.WaitGroup{}
	sem := make(chan struct{}, HasSpamMaxAsyncRequests)

	for inputData := range in {
		msgID, ok := inputData.(MsgID)
		if !ok {
			log.Println("В канал пришли данные, которые нельзя преобразовать к MsgID")
			continue
		}

		wg.Add(1)

		go func(msgID MsgID) {
			defer wg.Done()
			defer func() { <-sem }()

			sem <- struct{}{}
			isSpam, err := HasSpam(msgID)
			if err != nil {
				log.Println("Ошибка при проверке сообщения на спам")
				return
			}
			out <- MsgData{ID: msgID, HasSpam: isSpam}
		}(msgID)
	}

	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
	messages := []MsgData{}

	for msgData := range in {
		message, ok := msgData.(MsgData)
		if !ok {
			log.Println("В канал пришли данные, которые не удается преобразовать в MsgData")
			continue
		}

		messages = append(messages, message)
	}

	slices.SortFunc(messages, func(a, b MsgData) int {
		if a.HasSpam == b.HasSpam {
			if a.ID < b.ID {
				return -1
			} else {
				return 1
			}
		}
		if a.HasSpam {
			return -1
		}
		return 1
	})

	for _, message := range messages {
		out <- fmt.Sprintf("%t %d", message.HasSpam, message.ID)
	}
}
