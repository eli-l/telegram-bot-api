package tgbotapi

import (
	"fmt"
	"net/http"
	"time"
)

func ExampleNewBotAPI() {
	bot, err := NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	fmt.Printf("Authorized on account %s", bot.Self.UserName)
	u := NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		fmt.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyParameters.MessageID = update.Message.MessageID

		_, err := bot.Send(msg)
		if err != nil {
			panic(err)
		}
	}
}

func ExampleNewWebhook() {
	bot, err := NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	fmt.Printf("Authorized on account %s", bot.Self.UserName)

	wh, err := NewWebhookWithCert("https://www.google.com:8443/"+bot.Token, FilePath("cert.pem"))

	if err != nil {
		panic(err)
	}

	_, err = bot.Request(wh)

	if err != nil {
		panic(err)
	}

	info, err := bot.GetWebhookInfo()

	if err != nil {
		panic(err)
	}

	if info.LastErrorDate != 0 {
		fmt.Printf("failed to set webhook: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)

	for update := range updates {
		fmt.Printf("%+v\n", update)
	}
}

func ExampleWebhookHandler() {
	bot, err := NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	fmt.Printf("Authorized on account %s", bot.Self.UserName)

	wh, err := NewWebhookWithCert("https://www.google.com:8443/"+bot.Token, FilePath("cert.pem"))

	if err != nil {
		panic(err)
	}

	_, err = bot.Request(wh)
	if err != nil {
		panic(err)
	}
	info, err := bot.GetWebhookInfo()
	if err != nil {
		panic(err)
	}
	if info.LastErrorDate != 0 {
		fmt.Printf("[Telegram callback failed]%s", info.LastErrorMessage)
	}

	http.HandleFunc("/"+bot.Token, func(w http.ResponseWriter, r *http.Request) {
		update, err := bot.HandleUpdate(r)
		if err != nil {
			fmt.Printf("%+v\n", err.Error())
		} else {
			fmt.Printf("%+v\n", *update)
		}
	})

	go http.ListenAndServeTLS("0.0.0.0:8443", "./tests/cert.pem", "./tests/key.pem", nil)
}
