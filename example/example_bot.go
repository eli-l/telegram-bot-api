package example

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eli-l/telegram-bot-api/v7"
)

func ExampleNewBotAPI() {
	cfg := tgbotapi.NewBotConfig("MyAwesomeBotToken", tgbotapi.APIEndpoint, true)
	bot := tgbotapi.NewBot(cfg)

	if err := bot.Validate(); err != nil {
		panic(err)
	}

	usr, err := bot.GetMe()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Authorized on account %s", usr.UserName)

	updCh, err := tgbotapi.NewPollingHandler(bot, tgbotapi.NewUpdate(0)).
		InitUpdatesChannel()
	if err != nil {
		panic(err)
	}

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updCh.Clear()

	for update := range updCh {
		if update.Message == nil {
			continue
		}

		fmt.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyParameters.MessageID = update.Message.MessageID

		_, err := bot.Send(msg)
		if err != nil {
			panic(err)
		}
	}
}

func ExampleNewWebhook() {
	cfg := tgbotapi.NewBotConfig("MyAwesomeBotToken", tgbotapi.APIEndpoint, true)
	bot := tgbotapi.NewBot(cfg)

	if err := bot.Validate(); err != nil {
		panic(err)
	}

	usr, err := bot.GetMe()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Authorized on account %s", usr.UserName)

	wh, err := tgbotapi.NewWebhookWithCert("https://www.google.com:8443/"+cfg.GetToken(), tgbotapi.FilePath("cert.pem"))

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

	updCh := tgbotapi.NewWebhookHandler(bot).
		ListenForWebhook("/bot")

	go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)

	for update := range updCh {
		fmt.Printf("%+v\n", update)
	}
}

func ExampleWebhookHandler() {
	cfg := tgbotapi.NewBotConfig("MyAwesomeBotToken", tgbotapi.APIEndpoint, true)
	bot := tgbotapi.NewBot(cfg)

	if err := bot.Validate(); err != nil {
		panic(err)
	}

	usr, err := bot.GetMe()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Authorized on account %s", usr.UserName)

	wh, err := tgbotapi.NewWebhookWithCert("https://www.google.com:8443/"+cfg.GetToken(), tgbotapi.FilePath("cert.pem"))

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

	http.HandleFunc("/"+cfg.GetToken(), func(w http.ResponseWriter, r *http.Request) {
		update, err := tgbotapi.UnmarshalUpdate(r)
		if err != nil {
			fmt.Printf("%+v\n", err.Error())
		} else {
			fmt.Printf("%+v\n", *update)
		}
	})

	go http.ListenAndServeTLS("0.0.0.0:8443", "./tests/cert.pem", "./tests/key.pem", nil)
}

func ExampleInlineConfig() {
	cfg := tgbotapi.NewBotConfig("MyAwesomeBotToken", tgbotapi.APIEndpoint, true)
	bot := tgbotapi.NewBot(cfg)

	if err := bot.Validate(); err != nil {
		panic(err)
	}

	usr, err := bot.GetMe()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Authorized on account %s", usr.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tgbotapi.NewPollingHandler(bot, u).InitUpdatesChannel()
	if err != nil {
		panic(err)
	}

	for update := range updates {
		if update.InlineQuery == nil { // if no inline query, ignore it
			continue
		}

		article := tgbotapi.NewInlineQueryResultArticle(update.InlineQuery.ID, "Echo", update.InlineQuery.Query)
		article.Description = update.InlineQuery.Query

		inlineConf := tgbotapi.InlineConfig{
			InlineQueryID: update.InlineQuery.ID,
			IsPersonal:    true,
			CacheTime:     0,
			Results:       []interface{}{article},
		}

		if _, err := bot.Request(inlineConf); err != nil {
			fmt.Println(err)
		}
	}
}
