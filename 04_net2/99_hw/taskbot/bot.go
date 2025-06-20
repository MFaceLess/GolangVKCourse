package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

type BotData struct {
	Bot     *tgbotapi.BotAPI
	Updates tgbotapi.UpdatesChannel
}

const (
	tasksCommand    = "/tasks"
	newCommand      = "/new "
	assignCommand   = "/assign_"
	unassignCommand = "/unassign_"
	resolveCommand  = "/resolve_"
	myCommand       = "/my"
	ownerCommand    = "/owner"
)

var (
	BotToken       = ""
	FilePath       = "./config.json"
	WebhookURL     = "https://a04014544182ac09c198232f5d6b79a8.serveo.net"
	ServerPort     = 8081
	NumPoolWorkers = 100
)

func getUpdatesChanel() (*BotData, error) {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return nil, fmt.Errorf("NewBotAPI failed: %s", err)
	}

	bot.Debug = true

	wh, err := tgbotapi.NewWebhook(WebhookURL)
	if err != nil {
		return nil, fmt.Errorf("NewWebhook failed: %s", err)
	}

	if _, err = bot.Request(wh); err != nil {
		return nil, fmt.Errorf("SetWebhook failed: %s", err)
	}

	return &BotData{
		Bot:     bot,
		Updates: bot.ListenForWebhook("/"),
	}, nil
}

func worker(bot *tgbotapi.BotAPI, updateChan <-chan *tgbotapi.Update, wg *sync.WaitGroup) {
	defer wg.Done()

	for update := range updateChan {
		processingUserMessage(bot, update)
	}
}

func processingUserMessage(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	senderUsername := fmt.Sprintf("@%s", update.Message.From.UserName)

	usersChats.Lock()
	usersChats.UserChats[senderUsername] = update.Message.Chat.ID
	usersChats.Unlock()

	var messageToUser string

	switch {
	case update.Message.Text == tasksCommand:
		messageToUser = getTasks(senderUsername)

	case startsWith(update.Message.Text, newCommand):
		taskName := strings.Replace(update.Message.Text, newCommand, "", 1)
		messageToUser = createTask(taskName, senderUsername)

	case startsWith(update.Message.Text, assignCommand):
		IDStr := strings.Replace(update.Message.Text, assignCommand, "", 1)
		ID, err := strconv.Atoi(IDStr)
		if err == nil {
			messageToUser = assignTask(ID, senderUsername, bot)
		}

	case startsWith(update.Message.Text, unassignCommand):
		IDStr := strings.Replace(update.Message.Text, unassignCommand, "", 1)
		ID, err := strconv.Atoi(IDStr)
		if err == nil {
			messageToUser = unassignTask(ID, senderUsername, bot)
		}

	case startsWith(update.Message.Text, resolveCommand):
		IDStr := strings.Replace(update.Message.Text, resolveCommand, "", 1)
		ID, err := strconv.Atoi(IDStr)
		if err == nil {
			messageToUser = resolveTask(ID, senderUsername, bot)
		}

	case update.Message.Text == myCommand:
		messageToUser = getMyTasks(senderUsername)

	case update.Message.Text == ownerCommand:
		messageToUser = getOwnTasks(senderUsername)

	}

	_, err := bot.Send(tgbotapi.NewMessage(
		update.Message.Chat.ID,
		messageToUser,
	))

	if err != nil {
		log.Println(err)
	}
}

func startTaskBot(ctx context.Context) error {
	botData, err := getUpdatesChanel()
	if err != nil {
		return err
	}

	go func() {
		if err = http.ListenAndServe(fmt.Sprintf(":%d", ServerPort), nil); err != nil {
			log.Fatal(err)
			return
		}

		log.Println("gracefully stopped")
	}()

	wg := &sync.WaitGroup{}

	sizeBuffer := NumPoolWorkers
	updateChanel := make(chan *tgbotapi.Update, sizeBuffer)

	wg.Add(NumPoolWorkers)
	for i := 0; i < NumPoolWorkers; i++ {
		go worker(botData.Bot, updateChanel, wg)
	}

	for {
		select {
		case <-ctx.Done():
			close(updateChanel)
			wg.Wait()
			return nil

		case update, ok := <-botData.Updates:
			if !ok {
				close(updateChanel)
				wg.Wait()
				return nil
			}

			updateChanel <- &update
		}
	}
}

func main() {
	jsonConfig, err := SetConfig(FilePath)
	if err != nil {
		log.Fatalf("error read config: %s", err.Error())
	}

	BotToken = jsonConfig.token

	if err = startTaskBot(context.Background()); err != nil {
		log.Fatalf("error running bot: %s", err.Error())
	}
}
