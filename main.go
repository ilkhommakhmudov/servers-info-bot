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
var chatID int64 = -1001864417452
var services = map[string]string{
	"app.hippo.uz":       "https://app.hippo.uz/ping",
	"excel.app.hippo.uz": "https://app.hippo.uz/excel/ping",
	"hippo.sog.uz":       "https://hippo.sog.uz/api/ping",
}

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
}

func main() {
	var waitChan = make(chan int, 1)
	bot, _ = tgbotapi.NewBotAPI(token)
	bot.Debug = true

	go startListenUpdates()

	go func() {
		var client = &http.Client{Timeout: time.Second * 5}
		for {
			for serviceName, serviceUrl := range services {
				var request, _ = http.NewRequest(http.MethodGet, serviceUrl, nil)
				switch response, err := client.Do(request); {
				case err != nil:
					var message = fmt.Sprintf("%s: CONNECTION FAILED", serviceName)
					bot.Send(tgbotapi.NewMessage(chatID, message))

				case response.StatusCode != http.StatusOK:
					var message = fmt.Sprintf("%s: status code: %d", serviceName, response.StatusCode)
					bot.Send(tgbotapi.NewMessage(chatID, message))
				}
			}

			<-time.After(time.Second * 10)
		}
	}()
	<-waitChan
}
