package tests

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/eli-l/telegram-bot-api/v5"
)

const (
	ExistingDocumentFileID  = "BQACAgQAAxkBAAIBlWWq0525h50qLvTvedniXBoF-0cNAAJNFAACtIdZUaDyZwc4Cj8cNAQ"
	ExistingVoiceFileID     = "AwADAgADWQADjMcoCeul6r_q52IyAg"
	ExistingVideoFileID     = "BAADAgADZgADjMcoCav432kYe0FRAg"
	ExistingVideoNoteFileID = "DQADAgADdQAD70cQSUK41dLsRMqfAg"
	ExistingStickerFileID   = "BQADAgADcwADjMcoCbdl-6eB--YPAg"
)

var (
	TestToken        string
	Channel          string
	ChatID           int64
	SupergroupChatID int64
	ReplyToMessageID int
)

func init() {
	var err error
	TestToken = os.Getenv("TELEGRAM_TESTBOT_TOKEN")
	SupergroupChatID, err = strconv.ParseInt(os.Getenv("TELEGRAM_SUPERGROUP_CHAT_ID"), 10, 64)
	if err != nil {
		panic(err)
	}
	Channel = os.Getenv("TELEGRAM_CHANNEL")
	ChatID, err = strconv.ParseInt(os.Getenv("TELEGRAM_CHAT_ID"), 10, 64)
	if err != nil {
		panic(err)
	}
	ReplyToMessageID, err = strconv.Atoi(os.Getenv("TELEGRAM_REPLY_TO_MESSAGE_ID"))
	if err != nil {
		panic(err)
	}
}

type testLogger struct {
	t *testing.T
}

func (t testLogger) Println(v ...interface{}) {
	t.t.Log(v...)
}

func (t testLogger) Printf(format string, v ...interface{}) {
	t.t.Logf(format, v...)
}

func getBot(t *testing.T) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(TestToken)
	require.NoError(t, err)
	bot.Debug = true

	logger := testLogger{t}
	err = tgbotapi.SetLogger(logger)
	require.NoError(t, err)

	return bot, err
}

func TestNewBotAPI_notoken(t *testing.T) {
	_, err := tgbotapi.NewBotAPI("")
	require.Error(t, err)
}

func TestGetUpdates(t *testing.T) {
	bot, _ := getBot(t)

	u := tgbotapi.NewUpdate(0)

	up, err := bot.GetUpdates(u)
	require.NoError(t, err)
	require.NotNil(t, up)
}

func TestSendWithMessage(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithMessageReply(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ReplyParameters.MessageID = ReplyToMessageID
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithMessageForward(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewForward(ChatID, ChatID, ReplyToMessageID)
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestCopyMessage(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	message, err := bot.Send(msg)
	require.NoError(t, err)

	copyMessageConfig := tgbotapi.NewCopyMessage(SupergroupChatID, message.Chat.ID, message.MessageID)
	messageID, err := bot.CopyMessage(copyMessageConfig)
	require.NoError(t, err)
	require.NotEqual(t, message.MessageID, messageID.MessageID)
}

func TestSendWithNewPhoto(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewPhoto(ChatID, tgbotapi.FilePath("./image.jpg"))
	msg.Caption = "Test"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewPhotoWithFileBytes(t *testing.T) {
	bot, _ := getBot(t)

	data, _ := os.ReadFile("./image.jpg")
	b := tgbotapi.FileBytes{Name: "image.jpg", Bytes: data}

	msg := tgbotapi.NewPhoto(ChatID, b)
	msg.Caption = "Test"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewPhotoWithFileReader(t *testing.T) {
	bot, _ := getBot(t)

	f, _ := os.Open("./image.jpg")
	reader := tgbotapi.FileReader{Name: "image.jpg", Reader: f}

	msg := tgbotapi.NewPhoto(ChatID, reader)
	msg.Caption = "Test"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewPhotoReply(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewPhoto(ChatID, tgbotapi.FilePath("./image.jpg"))
	msg.ReplyParameters.MessageID = ReplyToMessageID

	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendNewPhotoToChannel(t *testing.T) {
	var photoID string
	bot, err := getBot(t)
	require.NoError(t, err)

	t.Run("send photo to channel", func(t *testing.T) {
		msg := tgbotapi.NewPhotoToChannel(Channel, tgbotapi.FilePath("./image.jpg"))
		msg.Caption = "Test"
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
		pl := len(m.Photo) > 0
		require.True(t, pl)
		photoID = m.Photo[0].FileID
	})

	t.Run("send photo to channel with existing photo", func(t *testing.T) {
		msg := tgbotapi.NewPhoto(ChatID, tgbotapi.FileID(photoID))
		msg.Caption = "Test existing"
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotEmpty(t, m)
	})

}

func TestSendNewPhotoToChannelFileBytes(t *testing.T) {
	bot, _ := getBot(t)

	data, _ := os.ReadFile("./image.jpg")
	b := tgbotapi.FileBytes{Name: "image.jpg", Bytes: data}

	msg := tgbotapi.NewPhotoToChannel(Channel, b)
	msg.Caption = "Test"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendNewPhotoToChannelFileReader(t *testing.T) {
	bot, _ := getBot(t)

	f, _ := os.Open("./image.jpg")
	reader := tgbotapi.FileReader{Name: "image.jpg", Reader: f}

	msg := tgbotapi.NewPhotoToChannel(Channel, reader)
	msg.Caption = "Test"
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)

}

func TestSendWithNewDocument(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewDocument(ChatID, tgbotapi.FilePath("./image.jpg"))
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewDocumentAndThumb(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewDocument(ChatID, tgbotapi.FilePath("./voice.ogg"))
	msg.Thumb = tgbotapi.FilePath("./image.jpg")
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.NotEmpty(t, m.Document.FileID)

}

func TestSendWithExistingDocument(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewDocument(ChatID, tgbotapi.FileID(ExistingDocumentFileID))
	m, err := bot.Send(msg)
	require.NotNil(t, m)
	require.NoError(t, err)
}

func TestSendWithAudio(t *testing.T) {
	var FileID string
	bot, err := getBot(t)
	require.NoError(t, err)

	t.Run("send new audio file", func(t *testing.T) {

		msg := tgbotapi.NewAudio(ChatID, tgbotapi.FilePath("./audio.mp3"))
		msg.Title = "TEST"
		msg.Duration = 10
		msg.Performer = "TEST"
		m, err := bot.Send(msg)
		require.NotNil(t, m)
		require.NoError(t, err)
		require.NotEmpty(t, m.Audio.FileID)
		FileID = m.Audio.FileID
	})

	t.Run("send existing audio file", func(t *testing.T) {
		msgExist := tgbotapi.NewAudio(ChatID, tgbotapi.FileID(FileID))
		msgExist.Title = "TEST EXIST"
		msgExist.Duration = 10
		msgExist.Performer = "TEST EXIST"
		m, err := bot.Send(msgExist)
		require.NotNil(t, m)
		require.NoError(t, err)
	})
}

func TestSendWithNewVoice(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewVoice(ChatID, tgbotapi.FilePath("./voice.ogg"))
	msg.Duration = 10
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

// TODO: fix this
//func TestSendWithExistingVoice(t *testing.T) {
//	bot, _ := getBot(t)
//
//	msg := tgbotapi.NewVoice(ChatID, tgbotapi.FileID(ExistingVoiceFileID))
//	msg.Duration = 10
//	_, err := bot.Send(msg)
//
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestSendWithContact(t *testing.T) {
	bot, _ := getBot(t)

	contact := tgbotapi.NewContact(ChatID, "5551234567", "Test")

	_, err := bot.Send(contact)
	require.NoError(t, err)
}

func TestSendWithLocation(t *testing.T) {
	bot, _ := getBot(t)

	_, err := bot.Send(tgbotapi.NewLocation(ChatID, 40, 40))
	require.NoError(t, err)
}

func TestSendWithVenue(t *testing.T) {
	bot, _ := getBot(t)

	venue := tgbotapi.NewVenue(ChatID, "A Test Location", "123 Test Street", 40, 40)

	_, err := bot.Send(venue)
	require.NoError(t, err)
}

func TestSendWithNewVideo(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewVideo(ChatID, tgbotapi.FilePath("./video.mp4"))
	msg.Duration = 10
	msg.Caption = "TEST"

	_, err := bot.Send(msg)
	require.NoError(t, err)
}

// TODO: fix this
//func TestSendWithExistingVideo(t *testing.T) {
//	bot, _ := getBot(t)
//
//	msg := tgbotapi.NewVideo(ChatID, tgbotapi.FileID(ExistingVideoFileID))
//	msg.Duration = 10
//	msg.Caption = "TEST"
//
//	_, err := bot.Send(msg)
//
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestSendWithNewVideoNote(t *testing.T) {
	bot, _ := getBot(t)
	msg := tgbotapi.NewVideoNote(ChatID, 240, tgbotapi.FilePath("./videonote.mp4"))
	msg.Duration = 10

	_, err := bot.Send(msg)
	require.NoError(t, err)
}

// TODO: fix this
//func TestSendWithExistingVideoNote(t *testing.T) {
//	bot, _ := getBot(t)
//
//	msg := tgbotapi.NewVideoNote(ChatID, 240, tgbotapi.FileID(ExistingVideoNoteFileID))
//	msg.Duration = 10
//
//	_, err := bot.Send(msg)
//
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestSendWithNewSticker(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewSticker(ChatID, tgbotapi.FilePath("./image.jpg"))

	_, err := bot.Send(msg)
	require.NoError(t, err)
}

// TODO: fix this
//func TestSendWithExistingSticker(t *testing.T) {
//	bot, _ := getBot(t)
//
//	msg := tgbotapi.NewSticker(ChatID, tgbotapi.FileID(ExistingStickerFileID))
//
//	_, err := bot.Send(msg)
//
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestSendWithNewStickerAndKeyboardHide(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewSticker(ChatID, tgbotapi.FilePath("./image.jpg"))
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
		RemoveKeyboard: true,
		Selective:      false,
	}
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

// TODO: fix this
//func TestSendWithExistingStickerAndKeyboardHide(t *testing.T) {
//	bot, _ := getBot(t)
//
//	msg := tgbotapi.NewSticker(ChatID, tgbotapi.FileID(ExistingStickerFileID))
//	msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
//		RemoveKeyboard: true,
//		Selective:      false,
//	}
//
//	_, err := bot.Send(msg)
//
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestSendWithDice(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewDice(ChatID)
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithDiceWithEmoji(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewDiceWithEmoji(ChatID, "🏀")
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

// TODO: fix this
//func TestGetFile(t *testing.T) {
//	bot, _ := getBot(t)
//
//	file := tgbotapi.FileConfig{
//		FileID: ExistingPhotoFileID,
//	}
//
//	_, err := bot.GetFile(file)
//
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestSendChatConfig(t *testing.T) {
	bot, _ := getBot(t)

	_, err := bot.Request(tgbotapi.NewChatAction(ChatID, tgbotapi.ChatTyping))
	require.NoError(t, err)
}

// TODO: identify why this isn't working
// func TestSendEditMessage(t *testing.T) {
// 	bot, _ := getBot(t)

// 	msg, err := bot.Send(NewMessage(ChatID, "Testing editing."))
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	edit := EditMessageTextConfig{
// 		BaseEdit: BaseEdit{
// 			ChatID:    ChatID,
// 			MessageID: msg.MessageID,
// 		},
// 		Text: "Updated text.",
// 	}

// 	_, err = bot.Send(edit)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

func TestGetUserProfilePhotos(t *testing.T) {
	bot, _ := getBot(t)

	_, err := bot.GetUserProfilePhotos(tgbotapi.NewUserProfilePhotos(ChatID))
	require.NoError(t, err)
}

func TestSetWebhookWithCert(t *testing.T) {
	bot, _ := getBot(t)

	time.Sleep(time.Second * 2)

	bot.Request(tgbotapi.DeleteWebhookConfig{})

	wh, err := tgbotapi.NewWebhookWithCert("https://example.com/tgbotapi-test/"+bot.Token, tgbotapi.FilePath("./cert.pem"))
	require.NoError(t, err)
	_, err = bot.Request(wh)
	require.NoError(t, err)

	_, err = bot.GetWebhookInfo()
	require.NoError(t, err)

	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
	require.NoError(t, err)
}

func TestSetWebhookWithoutCert(t *testing.T) {
	bot, _ := getBot(t)

	time.Sleep(time.Second * 2)

	bot.Request(tgbotapi.DeleteWebhookConfig{})

	wh, err := tgbotapi.NewWebhook("https://example.com/tgbotapi-test/" + bot.Token)
	require.NoError(t, err)

	_, err = bot.Request(wh)
	require.NoError(t, err)

	info, err := bot.GetWebhookInfo()
	require.NoError(t, err)
	require.NotEqual(t, 0, info.MaxConnections)
	require.Equal(t, 0, info.LastErrorDate)
	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
	require.NoError(t, err)
}

func TestSendWithMediaGroupPhotoVideo(t *testing.T) {
	bot, _ := getBot(t)

	cfg := tgbotapi.NewMediaGroup(ChatID, []interface{}{
		tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL("https://github.com/go-telegram-bot-api/telegram-bot-api/raw/0a3a1c8716c4cd8d26a262af9f12dcbab7f3f28c/tests/image.jpg")),
		tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath("./image.jpg")),
		tgbotapi.NewInputMediaVideo(tgbotapi.FilePath("./video.mp4")),
	})

	messages, err := bot.SendMediaGroup(cfg)
	require.NoError(t, err)
	require.NotNil(t, messages)
	require.Equal(t, len(cfg.Media), len(messages))
}

func TestSendWithMediaGroupDocument(t *testing.T) {
	bot, _ := getBot(t)

	cfg := tgbotapi.NewMediaGroup(ChatID, []interface{}{
		tgbotapi.NewInputMediaDocument(tgbotapi.FileURL("https://i.imgur.com/unQLJIb.jpg")),
		tgbotapi.NewInputMediaDocument(tgbotapi.FilePath("./image.jpg")),
	})

	messages, err := bot.SendMediaGroup(cfg)
	require.NoError(t, err)
	require.NotNil(t, messages)
	require.Equal(t, len(cfg.Media), len(messages))
}

func TestSendWithMediaGroupAudio(t *testing.T) {
	bot, _ := getBot(t)

	cfg := tgbotapi.NewMediaGroup(ChatID, []interface{}{
		tgbotapi.NewInputMediaAudio(tgbotapi.FilePath("./audio.mp3")),
		tgbotapi.NewInputMediaAudio(tgbotapi.FilePath("./audio.mp3")),
	})

	messages, err := bot.SendMediaGroup(cfg)
	require.NoError(t, err)
	require.NotNil(t, messages)
	require.Equal(t, len(cfg.Media), len(messages))
}

func ExampleNewBotAPI() {
	bot, err := tgbotapi.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	fmt.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
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

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyParameters.MessageID = update.Message.MessageID

		_, err := bot.Send(msg)
		if err != nil {
			panic(err)
		}
	}
}

func ExampleNewWebhook() {
	bot, err := tgbotapi.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	fmt.Printf("Authorized on account %s", bot.Self.UserName)

	wh, err := tgbotapi.NewWebhookWithCert("https://www.google.com:8443/"+bot.Token, tgbotapi.FilePath("cert.pem"))

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
	bot, err := tgbotapi.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	fmt.Printf("Authorized on account %s", bot.Self.UserName)

	wh, err := tgbotapi.NewWebhookWithCert("https://www.google.com:8443/"+bot.Token, tgbotapi.FilePath("cert.pem"))

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

	go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)
}

func ExampleInlineConfig() {
	bot, err := tgbotapi.NewBotAPI("MyAwesomeBotToken") // create new bot
	if err != nil {
		panic(err)
	}

	fmt.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

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

func TestDeleteMessage(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = tgbotapi.ModeMarkdown
	message, _ := bot.Send(msg)

	deleteMessageConfig := tgbotapi.DeleteMessageConfig{
		BaseChatMessage: tgbotapi.BaseChatMessage{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: message.Chat.ID,
			},
			MessageID: message.MessageID,
		},
	}
	_, err := bot.Request(deleteMessageConfig)

	if err != nil {
		t.Error(err)
	}
}

func TestPinChatMessage(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewMessage(SupergroupChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = tgbotapi.ModeMarkdown
	message, _ := bot.Send(msg)

	pinChatMessageConfig := tgbotapi.PinChatMessageConfig{
		BaseChatMessage: tgbotapi.BaseChatMessage{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: ChatID,
			},
			MessageID: message.MessageID,
		},
		DisableNotification: false,
	}
	_, err := bot.Request(pinChatMessageConfig)

	if err != nil {
		t.Error(err)
	}
}

func TestUnpinChatMessage(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewMessage(SupergroupChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = tgbotapi.ModeMarkdown
	message, err := bot.Send(msg)
	require.NoError(t, err)

	// We need pin message to unpin something
	pinChatMessageConfig := tgbotapi.PinChatMessageConfig{
		BaseChatMessage: tgbotapi.BaseChatMessage{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: message.Chat.ID,
			},
			MessageID: message.MessageID,
		},
		DisableNotification: false,
	}

	_, err = bot.Request(pinChatMessageConfig)
	require.NoError(t, err)

	unpinChatMessageConfig := tgbotapi.UnpinChatMessageConfig{
		BaseChatMessage: tgbotapi.BaseChatMessage{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: message.Chat.ID,
			},
			MessageID: message.MessageID,
		},
	}

	_, err = bot.Request(unpinChatMessageConfig)
	require.NoError(t, err)
}

func TestUnpinAllChatMessages(t *testing.T) {
	bot, _ := getBot(t)

	msg := tgbotapi.NewMessage(SupergroupChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = tgbotapi.ModeMarkdown
	message, _ := bot.Send(msg)

	pinChatMessageConfig := tgbotapi.PinChatMessageConfig{
		BaseChatMessage: tgbotapi.BaseChatMessage{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: message.Chat.ID,
			},
			MessageID: message.MessageID,
		},
		DisableNotification: true,
	}

	_, err := bot.Request(pinChatMessageConfig)
	require.NoError(t, err)

	unpinAllChatMessagesConfig := tgbotapi.UnpinAllChatMessagesConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: message.Chat.ID},
	}

	_, err = bot.Request(unpinAllChatMessagesConfig)
	require.NoError(t, err)
}

func TestPolls(t *testing.T) {
	bot, _ := getBot(t)

	poll := tgbotapi.NewPoll(SupergroupChatID, "Are polls working?", "Yes", "No")

	msg, err := bot.Send(poll)
	if err != nil {
		t.Error(err)
	}

	result, err := bot.StopPoll(tgbotapi.NewStopPoll(SupergroupChatID, msg.MessageID))
	if err != nil {
		t.Error(err)
	}

	if result.Question != "Are polls working?" {
		t.Error("Poll question did not match")
	}

	if !result.IsClosed {
		t.Error("Poll did not end")
	}

	if result.Options[0].Text != "Yes" || result.Options[0].VoterCount != 0 || result.Options[1].Text != "No" || result.Options[1].VoterCount != 0 {
		t.Error("Poll options were incorrect")
	}
}

// TODO: TG reports this as unsupported
func TestSendDice(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	dice := tgbotapi.NewDice(ChatID)

	msg, err := bot.Send(dice)
	require.NoError(t, err)
	require.NotNil(t, msg.Dice)
}

func TestCommands(t *testing.T) {
	bot, _ := getBot(t)

	setCommands := tgbotapi.NewSetMyCommands(tgbotapi.BotCommand{
		Command:     "test",
		Description: "a test command",
	})

	if _, err := bot.Request(setCommands); err != nil {
		t.Error("Unable to set commands")
	}

	commands, err := bot.GetMyCommands()
	if err != nil {
		t.Error("Unable to get commands")
	}

	if len(commands) != 1 {
		t.Error("Incorrect number of commands returned")
	}

	if commands[0].Command != "test" || commands[0].Description != "a test command" {
		t.Error("Commands were incorrectly set")
	}

	setCommands = tgbotapi.NewSetMyCommandsWithScope(tgbotapi.NewBotCommandScopeAllPrivateChats(), tgbotapi.BotCommand{
		Command:     "private",
		Description: "a private command",
	})

	if _, err := bot.Request(setCommands); err != nil {
		t.Error("Unable to set commands")
	}

	commands, err = bot.GetMyCommandsWithConfig(tgbotapi.NewGetMyCommandsWithScope(tgbotapi.NewBotCommandScopeAllPrivateChats()))
	if err != nil {
		t.Error("Unable to get commands")
	}

	if len(commands) != 1 {
		t.Error("Incorrect number of commands returned")
	}

	if commands[0].Command != "private" || commands[0].Description != "a private command" {
		t.Error("Commands were incorrectly set")
	}
}

// TODO: figure out why test is failing
//
// func TestEditMessageMedia(t *testing.T) {
// 	bot, _ := getBot(t)

// 	msg := NewPhoto(ChatID, "./image.jpg")
// 	msg.Caption = "Test"
// 	m, err := bot.Send(msg)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	edit := EditMessageMediaConfig{
// 		BaseEdit: BaseEdit{
// 			ChatID:    ChatID,
// 			MessageID: m.MessageID,
// 		},
// 		Media: NewInputMediaVideo(FilePath("./video.mp4")),
// 	}

// 	_, err = bot.Request(edit)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }
