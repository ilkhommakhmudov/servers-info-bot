package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"
)

var bot *tgbotapi.BotAPI
var token string
var administrationChatID int64 = -1001864417452 // admin tasks
var supportChatID int64 = -1002031508824

var notifierMap map[int64][]Service

func handleUpdate(update tgbotapi.Update) {
	// do nothing
}

func startListenUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 20

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		handleUpdate(update)
	}
}

func init() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err)
	}

	token = os.Getenv("TELEGRAM_BOT_TOKEN")

	notifierMap = make(map[int64][]Service)
	notifierMap[administrationChatID] = []Service{
		{
			CheckUrl: "https://app.hippo.uz/ping",
			Name:     "app.hippo.uz",
		},
		{
			CheckUrl: "https://hippo.sog.uz/api/ping",
			Name:     "hippo.sog.uz",
		},
	}
	notifierMap[supportChatID] = []Service{{
		CheckUrl: "https://app.hippo.uz/excel/ping",
		Name:     "app.hippo.uz - excel service",
	}}
}

func main() {
	var waitChan = make(chan int, 1)
	bot, _ = tgbotapi.NewBotAPI(token)
	bot.Debug = true

	go startListenUpdates()

	go func() {
		var client = &http.Client{Timeout: time.Second * 5}
		for {
			for chatID, services := range notifierMap {
				for _, service := range services {
					var request, _ = http.NewRequest(http.MethodGet, service.CheckUrl, nil)
					switch response, err := client.Do(request); {
					case err != nil:
						var message = fmt.Sprintf("%s: CONNECTION FAILED", service.Name)
						bot.Send(tgbotapi.NewMessage(chatID, message))

					case response.StatusCode != http.StatusOK:
						var message = fmt.Sprintf("%s: status code: %d", service.Name, response.StatusCode)
						bot.Send(tgbotapi.NewMessage(chatID, message))
					}
				}
			}
			<-time.After(time.Second * 10)
		}
	}()
	<-waitChan
}

type Service struct {
	CheckUrl string
	Name     string
}
