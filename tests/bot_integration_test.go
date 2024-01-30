package tests

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/eli-l/telegram-bot-api/v7"
)

var (
	TestToken        string
	Channel          string
	ChatID           int64
	SupergroupChatID int64
	ReplyToMessageID int
	Debug            = false
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
	bot := tgbotapi.NewBot(tgbotapi.NewBotConfig(TestToken, tgbotapi.APIEndpoint, Debug))
	err := bot.Validate()
	require.NoError(t, err)

	logger := testLogger{t}
	err = tgbotapi.SetLogger(logger)
	require.NoError(t, err)

	return bot, err
}

func TestNewBotAPI_notoken(t *testing.T) {
	bot := tgbotapi.NewBot(tgbotapi.NewBotConfig("", tgbotapi.APIEndpoint, Debug))
	require.NotNil(t, bot)
	err := bot.Validate()
	require.Error(t, err)
}

func TestGetUpdates(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)
	u := tgbotapi.NewUpdate(0)

	up, err := bot.GetUpdates(u)
	require.NoError(t, err)
	require.NotNil(t, up)
}

func TestSendWithMessage(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = tgbotapi.ModeMarkdown
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendWithMessageReply(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ReplyParameters.MessageID = ReplyToMessageID
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendWithMessageForward(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewForward(ChatID, ChatID, ReplyToMessageID)
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestCopyMessage(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	message, err := bot.Send(msg)
	require.NoError(t, err)

	copyMessageConfig := tgbotapi.NewCopyMessage(SupergroupChatID, message.Chat.ID, message.MessageID)
	messageID, err := bot.CopyMessage(copyMessageConfig)
	require.NoError(t, err)
	require.NotEqual(t, message.MessageID, messageID.MessageID)
}

func TestSendWithNewPhoto(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewPhoto(ChatID, tgbotapi.FilePath("./image.jpg"))
	msg.Caption = "Test"
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendWithNewPhotoWithFileBytes(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	data, err := os.ReadFile("./image.jpg")
	require.NoError(t, err)
	b := tgbotapi.FileBytes{Name: "image.jpg", Bytes: data}

	msg := tgbotapi.NewPhoto(ChatID, b)
	msg.Caption = "Test"
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendWithNewPhotoWithFileReader(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	f, err := os.Open("./image.jpg")
	require.NoError(t, err)
	reader := tgbotapi.FileReader{Name: "image.jpg", Reader: f}

	msg := tgbotapi.NewPhoto(ChatID, reader)
	msg.Caption = "Test"
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendWithNewPhotoReply(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewPhoto(ChatID, tgbotapi.FilePath("./image.jpg"))
	msg.ReplyParameters.MessageID = ReplyToMessageID

	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)
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
		require.NotEmpty(t, photoID)
		msg := tgbotapi.NewPhoto(ChatID, tgbotapi.FileID(photoID))
		msg.Caption = "Test existing"
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotEmpty(t, m)
	})

}

func TestSendNewPhotoToChannelFileBytes(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	data, err := os.ReadFile("./image.jpg")
	require.NoError(t, err)
	b := tgbotapi.FileBytes{Name: "image.jpg", Bytes: data}

	msg := tgbotapi.NewPhotoToChannel(Channel, b)
	msg.Caption = "Test"
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendNewPhotoToChannelFileReader(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	f, err := os.Open("./image.jpg")
	require.NoError(t, err)
	reader := tgbotapi.FileReader{Name: "image.jpg", Reader: f}

	msg := tgbotapi.NewPhotoToChannel(Channel, reader)
	msg.Caption = "Test"
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)

}

func TestSendWithNewDocument(t *testing.T) {
	var FileID string
	bot, err := getBot(t)
	require.NoError(t, err)

	t.Run("send new document", func(t *testing.T) {
		msg := tgbotapi.NewDocument(ChatID, tgbotapi.FilePath("./image.jpg"))
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
		require.NotEmpty(t, m.Document.FileID)
		FileID = m.Document.FileID
	})

	t.Run("get document", func(t *testing.T) {
		f, err := bot.GetFile(tgbotapi.FileConfig{FileID: FileID})
		require.NoError(t, err)
		require.NotNil(t, f)
		require.Equal(t, FileID, f.FileID)
	})

}

func TestSendWithNewDocumentAndThumb(t *testing.T) {
	var FileID string

	bot, err := getBot(t)
	require.NoError(t, err)

	t.Run("send new document and thumb", func(t *testing.T) {
		msg := tgbotapi.NewDocument(ChatID, tgbotapi.FilePath("./voice.ogg"))
		msg.Thumb = tgbotapi.FilePath("./image.jpg")
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
		require.NotEmpty(t, m.Document.FileID)
		FileID = m.Document.FileID
	})

	t.Run("send existing document", func(t *testing.T) {
		require.NotEmpty(t, FileID)
		msg := tgbotapi.NewDocument(ChatID, tgbotapi.FileID(FileID))
		m, err := bot.Send(msg)
		require.NotNil(t, m)
		require.NoError(t, err)
	})

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
		require.NotEmpty(t, FileID)
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
	var FileID string

	bot, err := getBot(t)
	require.NoError(t, err)

	t.Run("send new voice file", func(t *testing.T) {
		msg := tgbotapi.NewVoice(ChatID, tgbotapi.FilePath("./voice.ogg"))
		msg.Duration = 10
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
		require.NotEmpty(t, m.Voice.FileID)
		FileID = m.Voice.FileID
	})

	t.Run("send existing voice file", func(t *testing.T) {
		require.NotEmpty(t, FileID)
		msg := tgbotapi.NewVoice(ChatID, tgbotapi.FileID(FileID))
		msg.Duration = 10
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
	})
}

func TestSendWithContact(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	contact := tgbotapi.NewContact(ChatID, "5551234567", "Test")

	m, err := bot.Send(contact)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendWithLocation(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	m, err := bot.Send(tgbotapi.NewLocation(ChatID, 40, 40))
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendWithVenue(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	venue := tgbotapi.NewVenue(ChatID, "A Test Location", "123 Test Street", 40, 40)

	m, err := bot.Send(venue)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendWithNewVideo(t *testing.T) {
	var FileID string

	bot, err := getBot(t)
	require.NoError(t, err)

	t.Run("send new video file", func(t *testing.T) {
		msg := tgbotapi.NewVideo(ChatID, tgbotapi.FilePath("./video.mp4"))
		msg.Duration = 10
		msg.Caption = "TEST"

		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
		require.NotEmpty(t, m.Video.FileID)
		FileID = m.Video.FileID
	})

	t.Run("send existing video file", func(t *testing.T) {
		require.NotEmpty(t, FileID)
		msg := tgbotapi.NewVideo(ChatID, tgbotapi.FileID(FileID))
		msg.Duration = 10
		msg.Caption = "TEST EXIST"
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
	})
}

func TestSendWithNewVideoNote(t *testing.T) {
	var FileID string

	bot, err := getBot(t)
	require.NoError(t, err)

	t.Run("send new video note file", func(t *testing.T) {
		msg := tgbotapi.NewVideoNote(ChatID, 240, tgbotapi.FilePath("./videonote.mp4"))
		msg.Duration = 10

		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotEmpty(t, m.VideoNote.FileID)
		FileID = m.VideoNote.FileID
	})

	t.Run("send existing video note file", func(t *testing.T) {
		require.NotEmpty(t, FileID)
		msg := tgbotapi.NewVideoNote(ChatID, 240, tgbotapi.FileID(FileID))
		msg.Duration = 10
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
	})
}

func TestSendWithNewSticker(t *testing.T) {
	var FileID string

	bot, err := getBot(t)
	require.NoError(t, err)

	t.Run("send new sticker file", func(t *testing.T) {
		msg := tgbotapi.NewSticker(ChatID, tgbotapi.FilePath("./image.jpg"))
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
		require.NotEmpty(t, m.Sticker.FileID)
		FileID = m.Sticker.FileID
	})

	t.Run("send existing sticker file", func(t *testing.T) {
		require.NotEmpty(t, FileID)
		msg := tgbotapi.NewSticker(ChatID, tgbotapi.FileID(FileID))
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
	})
}

func TestSendWithNewStickerAndKeyboardHide(t *testing.T) {
	var FileID string

	bot, err := getBot(t)
	require.NoError(t, err)

	t.Run("send new sticker file", func(t *testing.T) {
		msg := tgbotapi.NewSticker(ChatID, tgbotapi.FilePath("./image.jpg"))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
			RemoveKeyboard: true,
			Selective:      false,
		}
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
		require.NotEmpty(t, m.Sticker.FileID)
		FileID = m.Sticker.FileID
	})

	t.Run("send existing sticker file", func(t *testing.T) {
		require.NotEmpty(t, FileID)
		msg := tgbotapi.NewSticker(ChatID, tgbotapi.FileID(FileID))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
			RemoveKeyboard: true,
			Selective:      false,
		}
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)
	})
}

func TestSendWithDice(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewDice(ChatID)
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendWithDiceWithEmoji(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewDiceWithEmoji(ChatID, "🏀")
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendChatConfig(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	m, err := bot.Request(tgbotapi.NewChatAction(ChatID, tgbotapi.ChatTyping))
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestSendEditMessage(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg, err := bot.Send(tgbotapi.NewMessage(ChatID, "Testing editing."))
	require.NoError(t, err)
	require.NotNil(t, msg)

	edit := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			BaseChatMessage: tgbotapi.BaseChatMessage{
				MessageID: msg.MessageID,
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: ChatID,
				},
			},
		},
		Text: "Updated text.",
	}

	m, err := bot.Send(edit)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestGetUserProfilePhotos(t *testing.T) {
	bot, _ := getBot(t)

	_, err := bot.GetUserProfilePhotos(tgbotapi.NewUserProfilePhotos(ChatID))
	require.NoError(t, err)
}

func TestSetWebhookWithCert(t *testing.T) {

	bot, _ := getBot(t)

	time.Sleep(time.Second * 2)

	bot.Request(tgbotapi.DeleteWebhookConfig{})

	wh, err := tgbotapi.NewWebhookWithCert("https://example.com/tgbotapi-test/"+bot.GetConfig().GetToken(), tgbotapi.FilePath("./cert.pem"))
	require.NoError(t, err)
	_, err = bot.Request(wh)
	require.NoError(t, err)

	_, err = bot.GetWebhookInfo()
	require.NoError(t, err)

	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
	require.NoError(t, err)
}

func TestSetWebhookWithoutCert(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	time.Sleep(time.Second * 2)

	bot.Request(tgbotapi.DeleteWebhookConfig{})

	wh, err := tgbotapi.NewWebhook("https://example.com/tgbotapi-test/" + bot.GetConfig().GetToken())
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
	bot, err := getBot(t)
	require.NoError(t, err)

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
	bot, err := getBot(t)
	require.NoError(t, err)

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
	bot, err := getBot(t)
	require.NoError(t, err)

	cfg := tgbotapi.NewMediaGroup(ChatID, []interface{}{
		tgbotapi.NewInputMediaAudio(tgbotapi.FilePath("./audio.mp3")),
		tgbotapi.NewInputMediaAudio(tgbotapi.FilePath("./audio.mp3")),
	})

	messages, err := bot.SendMediaGroup(cfg)
	require.NoError(t, err)
	require.NotNil(t, messages)
	require.Equal(t, len(cfg.Media), len(messages))
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
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewMessage(SupergroupChatID, "Test message. It is supposed to be pinned after test is done. From telegram-bot-api.")
	msg.ParseMode = tgbotapi.ModeMarkdown
	message, err := bot.Send(msg)
	require.NoError(t, err)

	pinChatMessageConfig := tgbotapi.PinChatMessageConfig{
		BaseChatMessage: tgbotapi.BaseChatMessage{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: message.Chat.ID,
			},
			MessageID: message.MessageID,
		},
		DisableNotification: false,
	}
	res, err := bot.Request(pinChatMessageConfig)
	require.NoError(t, err)
	require.NotNil(t, res)
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

func TestSendDice(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	dice := tgbotapi.NewDice(ChatID)

	msg, err := bot.Send(dice)
	require.NoError(t, err)
	require.NotNil(t, msg.Dice)
}

func TestCommands(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	setCommands := tgbotapi.NewSetMyCommands(tgbotapi.BotCommand{
		Command:     "test",
		Description: "a test command",
	})

	_, err = bot.Request(setCommands)
	require.NoError(t, err)

	commands, err := bot.GetMyCommands()
	require.NoError(t, err)
	require.Equal(t, 1, len(commands))
	require.Equal(t, "test", commands[0].Command)
	require.Equal(t, "a test command", commands[0].Description)

	setCommands = tgbotapi.NewSetMyCommandsWithScope(tgbotapi.NewBotCommandScopeAllPrivateChats(), tgbotapi.BotCommand{
		Command:     "private",
		Description: "a private command",
	})

	_, err = bot.Request(setCommands)
	require.NoError(t, err)

	commands, err = bot.GetMyCommandsWithConfig(tgbotapi.NewGetMyCommandsWithScope(tgbotapi.NewBotCommandScopeAllPrivateChats()))
	require.NoError(t, err)
	require.Equal(t, 1, len(commands))
	require.Equal(t, "private", commands[0].Command)
	require.Equal(t, "a private command", commands[0].Description)
}

func TestEditMessageMedia(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	msg := tgbotapi.NewPhoto(ChatID, tgbotapi.FilePath("./image.jpg"))
	msg.Caption = "Test"
	m, err := bot.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, m)

	edit := tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			BaseChatMessage: tgbotapi.BaseChatMessage{
				MessageID: m.MessageID,
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: ChatID,
				},
			},
		},
		Media: tgbotapi.NewInputMediaVideo(tgbotapi.FilePath("./video.mp4")),
	}

	res, err := bot.Request(edit)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestSetReaction(t *testing.T) {
	bot, err := getBot(t)
	require.NoError(t, err)

	t.Run("set reaction using reaction type", func(t *testing.T) {
		msg := tgbotapi.NewMessage(ChatID, "An initial message to test reaction type")
		msg.ParseMode = tgbotapi.ModeMarkdown
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)

		reaction := tgbotapi.NewSetMessageReactionType(ChatID, m.MessageID, tgbotapi.ReactionType{
			Type:  "emoji",
			Emoji: "👍",
		}, true)

		res, err := bot.Request(reaction)
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("set reaction using reaction emoji", func(t *testing.T) {
		msg := tgbotapi.NewMessage(ChatID, "An initial message to test reaction emoji")
		msg.ParseMode = tgbotapi.ModeMarkdown
		m, err := bot.Send(msg)
		require.NoError(t, err)
		require.NotNil(t, m)

		reaction := tgbotapi.NewSetMessageReactionEmoji(ChatID, m.MessageID, "👀", true)

		res, err := bot.Request(reaction)
		require.NoError(t, err)
		require.NotNil(t, res)
	})
}
