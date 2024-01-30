package tgbotapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	TestToken        = "ThisIsATestTokenNoMatterWhatItContains"
	ChatID           = 111
	SupergroupChatID = -1111
	ReplyToMessageID = 1
)

func prepareHttpClient(t *testing.T) *MockHTTPClient {
	ctrl := gomock.NewController(t)
	httpMock := NewMockHTTPClient(ctrl)
	return httpMock
}

func expectGetMe(t *testing.T, c *MockHTTPClient) {
	meResp := `{"ok": true, "result": {"id": 123456789, "is_bot": true, "first_name": "MyBot", "username": "my_bot"}}`
	c.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			valid := isRequestValid(req, TestToken, "getMe")
			require.True(t, valid)
			return newOKResponse(meResp), nil
		})
}

func TestNewBotInstance_WithEmptyToken(t *testing.T) {
	bot := NewBot(NewDefaultBotConfig(""))
	err := bot.Validate()
	require.Error(t, err)
}

func isRequestValid(r *http.Request, token, path string) bool {
	return r.Method == http.MethodPost &&
		r.URL.Path == fmt.Sprintf("/bot%s/%s", token, path) &&
		r.URL.Scheme == "https"
}

func newOKResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}

func TestGetUpdates(t *testing.T) {
	updateResp := `{
        "ok": true,
        "result": [
            {
                "update_id": 123456789,
                "message": {
                    "message_id": 111,
                    "from": {
                        "id": 222,
                        "is_bot": false,
                        "first_name": "John",
                        "username": "john_doe"
                    },
                    "chat": {
                        "id": 333,
                        "first_name": "John",
                        "username": "john_doe",
                        "type": "private"
                    },
                    "date": 1640001112,
                    "text": "Hello, bot!"
                }
            }
        ]
    }`

	client := prepareHttpClient(t)
	defer client.ctrl.Finish()
	expectGetMe(t, client)

	client.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			valid := isRequestValid(req, TestToken, "getUpdates")
			require.True(t, valid)
			return newOKResponse(updateResp), nil
		})

	bot := NewBotWithClient(NewBotConfig(TestToken, APIEndpoint, false), client)
	err := bot.Validate()
	require.NoError(t, err)

	u := NewUpdate(0)
	_, err = bot.GetUpdates(u)
	require.NoError(t, err)
}

func TestSendWithMessage(t *testing.T) {
	client := prepareHttpClient(t)
	defer client.ctrl.Finish()

	responseBody := `{
	       "ok": true,
	       "result": {
	           "message_id": 123,
	           "from": {
	               "id": 123456789,
	               "is_bot": true,
	               "first_name": "MyBot",
	               "username": "my_bot"
	           },
	           "chat": {
	               "id": 987654321,
	               "first_name": "John",
	               "username": "john_doe",
	               "type": "private"
	           },
	           "date": 1640001112,
	           "text": "Hello, John!"
	       }
	   }`

	expectGetMe(t, client)

	client.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			valid := isRequestValid(req, TestToken, "sendMessage")
			require.True(t, valid)
			return newOKResponse(responseBody), nil
		})

	bot := NewBotWithClient(NewBotConfig(TestToken, APIEndpoint, false), client)
	err := bot.Validate()
	require.NoError(t, err)

	msg := NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = ModeMarkdown
	_, err = bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithMessageReply(t *testing.T) {
	client := prepareHttpClient(t)
	defer client.ctrl.Finish()

	responseBody := `{
	       "ok": true,
	       "result": {
	           "message_id": 123,
	           "from": {
	               "id": 123456789,
	               "is_bot": true,
	               "first_name": "MyBot",
	               "username": "my_bot"
	           },
	           "chat": {
	               "id": 987654321,
	               "first_name": "John",
	               "username": "john_doe",
	               "type": "private"
	           },
	           "date": 1640001112,
	           "text": "Hello, John!"
	       }
	   }`

	expectGetMe(t, client)

	client.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			valid := isRequestValid(req, TestToken, "sendMessage")
			require.True(t, valid)
			return newOKResponse(responseBody), nil
		})

	bot := NewBotWithClient(NewBotConfig(TestToken, APIEndpoint, false), client)
	err := bot.Validate()
	require.NoError(t, err)

	msg := NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ReplyParameters.MessageID = ReplyToMessageID
	_, err = bot.Send(msg)
	require.NoError(t, err)
}

func TestCopyMessage(t *testing.T) {
	client := prepareHttpClient(t)
	defer client.ctrl.Finish()
	expectGetMe(t, client)

	responseBody := `{
	       "ok": true,
	       "result": {
	           "message_id": 123,
	           "from": {
	               "id": 123456789,
	               "is_bot": true,
	               "first_name": "MyBot",
	               "username": "my_bot"
	           },
	           "chat": {
	               "id": 987654321,
	               "first_name": "John",
	               "username": "john_doe",
	               "type": "private"
	           },
	           "date": 1640001112,
	           "text": "Hello, John!"
	       }
	   }`

	client.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			valid := isRequestValid(req, TestToken, "sendMessage")
			require.True(t, valid)
			return newOKResponse(responseBody), nil
		})

	client.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			valid := isRequestValid(req, TestToken, "copyMessage")
			require.True(t, valid)
			return newOKResponse(responseBody), nil
		})

	bot := NewBotWithClient(NewBotConfig(TestToken, APIEndpoint, false), client)
	err := bot.Validate()
	require.NoError(t, err)

	msg := NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	message, err := bot.Send(msg)
	require.NoError(t, err)

	copyMessageConfig := NewCopyMessage(SupergroupChatID, message.Chat.ID, message.MessageID)
	messageID, err := bot.CopyMessage(copyMessageConfig)
	require.NoError(t, err)

	require.Equal(t, messageID.MessageID, message.MessageID)
}
func TestPrepareInputMediaForParams(t *testing.T) {
	media := []any{
		NewInputMediaPhoto(FilePath("./image.jpg")),
		NewInputMediaVideo(FileID("test")),
	}

	prepared := prepareInputMediaForParams(media)

	if media[0].(InputMediaPhoto).Media != FilePath("./image.jpg") {
		t.Error("Original media was changed")
	}

	if prepared[0].(InputMediaPhoto).Media != fileAttach("attach://file-0") {
		t.Error("New media was not replaced")
	}

	if prepared[1].(InputMediaVideo).Media != FileID("test") {
		t.Error("Passthrough value was not the same")
	}
}
