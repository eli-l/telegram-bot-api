// Package tgbotapi has functions and types used for interacting with
// the Telegram Bot API.
package tgbotapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

const DefaultBufferSize = 100

// HTTPClientI is the type needed for the bot to perform HTTP requests.
type HTTPClientI interface {
	Do(req *http.Request) (*http.Response, error)
}

// BotAPI allows you to interact with the Telegram Bot API.
type BotAPI struct {
	config   BotConfigI
	client   HTTPClientI
	readonly bool
}

// NewBot creates a new BotAPI instance.
//
// It requires a token, provided by @BotFather on Telegram.
func NewBot(config BotConfigI) *BotAPI {
	return NewBotWithClient(config, &http.Client{})
}

// NewBotWithClient creates a new BotAPI instance
// and allows you to pass a http.Client.
//
// It requires a token, provided by @BotFather on Telegram and API endpoint.
func NewBotWithClient(config BotConfigI, client HTTPClientI) *BotAPI {
	bot := &BotAPI{
		config: config.(*BotConfig),
		client: client,
	}

	return bot
}

func (b *BotAPI) GetConfig() BotConfigI {
	return b.config
}

func (b *BotAPI) Validate() error {
	_, err := b.GetMe()
	return err
}

func (bot *BotAPI) SetReadonly(readonly bool) {
	bot.readonly = readonly
}

func buildParams(in Params) url.Values {
	if in == nil {
		return url.Values{}
	}

	out := url.Values{}

	for key, value := range in {
		out.Set(key, value)
	}

	return out
}

// MakeRequest makes a request to a specific endpoint with our token.
func (bot *BotAPI) MakeRequest(endpoint string, params Params) (*APIResponse, error) {
	if bot.config.GetDebug() {
		log.Printf("Endpoint: %s, params: %v\n", endpoint, params)
	}

	method := fmt.Sprintf(bot.config.GetApiEndpoint(), bot.config.GetToken(), endpoint)

	values := buildParams(params)

	req, err := http.NewRequest("POST", method, strings.NewReader(values.Encode()))
	if err != nil {
		return &APIResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := bot.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	bytes, err := bot.decodeAPIResponse(resp.Body, &apiResp)
	if err != nil {
		return &apiResp, err
	}

	if bot.config.GetDebug() {
		log.Printf("Endpoint: %s, response: %s\n", endpoint, string(bytes))
	}

	if !apiResp.Ok {
		var parameters ResponseParameters

		if apiResp.Parameters != nil {
			parameters = *apiResp.Parameters
		}

		return &apiResp, &Error{
			Code:               apiResp.ErrorCode,
			Message:            apiResp.Description,
			ResponseParameters: parameters,
		}
	}

	return &apiResp, nil
}

// decodeAPIResponse decode response and return slice of bytes if debug enabled.
// If debug disabled, just decode http.Response.Body stream to APIResponse struct
// for efficient memory usage
func (bot *BotAPI) decodeAPIResponse(responseBody io.Reader, resp *APIResponse) ([]byte, error) {
	if !bot.config.GetDebug() {
		dec := json.NewDecoder(responseBody)
		err := dec.Decode(resp)
		return nil, err
	}

	// if debug, read response body
	data, err := io.ReadAll(responseBody)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// UploadFiles makes a request to the API with files.
func (bot *BotAPI) UploadFiles(endpoint string, params Params, files []RequestFile) (*APIResponse, error) {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	// This code modified from the very helpful @HirbodBehnam
	// https://github.com/go-telegram-bot-api/telegram-bot-api/issues/354#issuecomment-663856473
	go func() {
		defer w.Close()
		defer m.Close()

		for field, value := range params {
			if err := m.WriteField(field, value); err != nil {
				w.CloseWithError(err)
				return
			}
		}

		for _, file := range files {
			if file.Data.NeedsUpload() {
				name, reader, err := file.Data.UploadData()
				if err != nil {
					w.CloseWithError(err)
					return
				}

				part, err := m.CreateFormFile(file.Name, name)
				if err != nil {
					w.CloseWithError(err)
					return
				}

				if _, err := io.Copy(part, reader); err != nil {
					w.CloseWithError(err)
					return
				}

				if closer, ok := reader.(io.ReadCloser); ok {
					if err = closer.Close(); err != nil {
						w.CloseWithError(err)
						return
					}
				}
			} else {
				value := file.Data.SendData()

				if err := m.WriteField(file.Name, value); err != nil {
					w.CloseWithError(err)
					return
				}
			}
		}
	}()

	if bot.config.GetDebug() {
		log.Printf("Endpoint: %s, params: %v, with %d files\n", endpoint, params, len(files))
	}

	method := fmt.Sprintf(bot.config.GetApiEndpoint(), bot.config.GetToken(), endpoint)

	req, err := http.NewRequest("POST", method, r)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", m.FormDataContentType())

	resp, err := bot.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	bytes, err := bot.decodeAPIResponse(resp.Body, &apiResp)
	if err != nil {
		return &apiResp, err
	}

	if bot.config.GetDebug() {
		log.Printf("Endpoint: %s, response: %s\n", endpoint, string(bytes))
	}

	if !apiResp.Ok {
		var parameters ResponseParameters

		if apiResp.Parameters != nil {
			parameters = *apiResp.Parameters
		}

		return &apiResp, &Error{
			Message:            apiResp.Description,
			ResponseParameters: parameters,
		}
	}

	return &apiResp, nil
}

// GetFileDirectURL returns direct URL to file
//
// It requires the FileID.
func (bot *BotAPI) GetFileDirectURL(fileID string) (string, error) {
	file, err := bot.GetFile(FileConfig{fileID})

	if err != nil {
		return "", err
	}

	return file.Link(bot.config.GetToken()), nil
}

// GetMe fetches the currently authenticated bot.
//
// This method is called upon creation to validate the token,
// and so you may get this data from BotAPI.Self without the need for
// another request.
func (bot *BotAPI) GetMe() (User, error) {
	resp, err := bot.MakeRequest("getMe", nil)
	if err != nil {
		return User{}, err
	}

	var user User
	err = json.Unmarshal(resp.Result, &user)

	return user, err
}

//// IsMessageToMe returns true if message directed to this bot.
////
//// It requires the Message.
//func (bot *BotAPI) IsMessageToMe(message Message) bool {
//	return strings.Contains(message.Text, "@"+bot.Self.UserName)
//}

func hasFilesNeedingUpload(files []RequestFile) bool {
	for _, file := range files {
		if file.Data.NeedsUpload() {
			return true
		}
	}

	return false
}

// Request sends a Chattable to Telegram, and returns the APIResponse.
func (bot *BotAPI) Request(c Chattable) (*APIResponse, error) {
	params, err := c.params()
	if err != nil {
		return nil, err
	}

	if t, ok := c.(Fileable); ok {
		files := t.files()

		// If we have files that need to be uploaded, we should delegate the
		// request to UploadFile.
		if hasFilesNeedingUpload(files) {
			return bot.UploadFiles(t.method(), params, files)
		}

		// However, if there are no files to be uploaded, there's likely things
		// that need to be turned into params instead.
		for _, file := range files {
			params[file.Name] = file.Data.SendData()
		}
	}

	return bot.MakeRequest(c.method(), params)
}

// Send will send a Chattable item to Telegram and provides the
// returned Message.
func (bot *BotAPI) Send(c Chattable) (Message, error) {
	resp, err := bot.Request(c)
	if err != nil {
		return Message{}, err
	}

	var message Message
	err = json.Unmarshal(resp.Result, &message)

	return message, err
}

// SendMediaGroup sends a media group and returns the resulting messages.
func (bot *BotAPI) SendMediaGroup(config MediaGroupConfig) ([]Message, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return nil, err
	}

	var messages []Message
	err = json.Unmarshal(resp.Result, &messages)

	return messages, err
}

// GetUserProfilePhotos gets a user's profile photos.
//
// It requires UserID.
// Offset and Limit are optional.
func (bot *BotAPI) GetUserProfilePhotos(config UserProfilePhotosConfig) (UserProfilePhotos, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return UserProfilePhotos{}, err
	}

	var profilePhotos UserProfilePhotos
	err = json.Unmarshal(resp.Result, &profilePhotos)

	return profilePhotos, err
}

// GetFile returns a File which can download a file from Telegram.
//
// Requires FileID.
func (bot *BotAPI) GetFile(config FileConfig) (File, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return File{}, err
	}

	var file File
	err = json.Unmarshal(resp.Result, &file)

	return file, err
}

// GetWebhookInfo allows you to fetch information about a webhook and if
// one currently is set, along with pending update count and error messages.
func (bot *BotAPI) GetWebhookInfo() (WebhookInfo, error) {
	resp, err := bot.MakeRequest("getWebhookInfo", nil)
	if err != nil {
		return WebhookInfo{}, err
	}

	var info WebhookInfo
	err = json.Unmarshal(resp.Result, &info)

	return info, err
}

// GetUpdates fetches updates.
// If a WebHook is set, this will not return any data!
//
// Offset, Limit, Timeout, and AllowedUpdates are optional.
// To avoid stale items, set Offset to one higher than the previous item.
// Set Timeout to a large number to reduce requests, so you can get updates
// instantly instead of having to wait between requests.
func (bot *BotAPI) GetUpdates(config UpdateConfig) ([]Update, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return []Update{}, err
	}

	var updates []Update
	err = json.Unmarshal(resp.Result, &updates)

	return updates, err
}

// WriteToHTTPResponse writes the request to the HTTP ResponseWriter.
//
// It doesn't support uploading files.
//
// See https://core.telegram.org/bots/api#making-requests-when-getting-updates
// for details.
func WriteToHTTPResponse(w http.ResponseWriter, c Chattable) error {
	params, err := c.params()
	if err != nil {
		return err
	}

	if t, ok := c.(Fileable); ok {
		if hasFilesNeedingUpload(t.files()) {
			return errors.New("unable to use http response to upload files")
		}
	}

	values := buildParams(params)
	values.Set("method", c.method())

	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	_, err = w.Write([]byte(values.Encode()))
	return err
}

// GetChat gets information about a chat.
func (bot *BotAPI) GetChat(config ChatInfoConfig) (ChatFullInfo, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return ChatFullInfo{}, err
	}

	var chat ChatFullInfo
	err = json.Unmarshal(resp.Result, &chat)

	return chat, err
}

// GetChatAdministrators gets a list of administrators in the chat.
//
// If none have been appointed, only the creator will be returned.
// Bots are not shown, even if they are an administrator.
func (bot *BotAPI) GetChatAdministrators(config ChatAdministratorsConfig) ([]ChatMember, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return []ChatMember{}, err
	}

	var members []ChatMember
	err = json.Unmarshal(resp.Result, &members)

	return members, err
}

// GetChatMembersCount gets the number of users in a chat.
func (bot *BotAPI) GetChatMembersCount(config ChatMemberCountConfig) (int, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return -1, err
	}

	var count int
	err = json.Unmarshal(resp.Result, &count)

	return count, err
}

// GetChatMember gets a specific chat member.
func (bot *BotAPI) GetChatMember(config GetChatMemberConfig) (ChatMember, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return ChatMember{}, err
	}

	var member ChatMember
	err = json.Unmarshal(resp.Result, &member)

	return member, err
}

// GetGameHighScores allows you to get the high scores for a game.
func (bot *BotAPI) GetGameHighScores(config GetGameHighScoresConfig) ([]GameHighScore, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return []GameHighScore{}, err
	}

	var highScores []GameHighScore
	err = json.Unmarshal(resp.Result, &highScores)

	return highScores, err
}

// GetInviteLink get InviteLink for a chat
func (bot *BotAPI) GetInviteLink(config ChatInviteLinkConfig) (string, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return "", err
	}

	var inviteLink string
	err = json.Unmarshal(resp.Result, &inviteLink)

	return inviteLink, err
}

// GetStickerSet returns a StickerSet.
func (bot *BotAPI) GetStickerSet(config GetStickerSetConfig) (StickerSet, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return StickerSet{}, err
	}

	var stickerSet StickerSet
	err = json.Unmarshal(resp.Result, &stickerSet)

	return stickerSet, err
}

// GetCustomEmojiStickers returns a slice of Sticker objects.
func (bot *BotAPI) GetCustomEmojiStickers(config GetCustomEmojiStickersConfig) ([]Sticker, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return []Sticker{}, err
	}

	var stickers []Sticker
	err = json.Unmarshal(resp.Result, &stickers)

	return stickers, err
}

// StopPoll stops a poll and returns the result.
func (bot *BotAPI) StopPoll(config StopPollConfig) (Poll, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return Poll{}, err
	}

	var poll Poll
	err = json.Unmarshal(resp.Result, &poll)

	return poll, err
}

// GetMyCommands gets the currently registered commands.
func (bot *BotAPI) GetMyCommands() ([]BotCommand, error) {
	return bot.GetMyCommandsWithConfig(GetMyCommandsConfig{})
}

// GetMyCommandsWithConfig gets the currently registered commands with a config.
func (bot *BotAPI) GetMyCommandsWithConfig(config GetMyCommandsConfig) ([]BotCommand, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return nil, err
	}

	var commands []BotCommand
	err = json.Unmarshal(resp.Result, &commands)

	return commands, err
}

// CopyMessage copy messages of any kind. The method is analogous to the method
// forwardMessage, but the copied message doesn't have a link to the original
// message. Returns the MessageID of the sent message on success.
func (bot *BotAPI) CopyMessage(config CopyMessageConfig) (MessageID, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return MessageID{}, err
	}

	var messageID MessageID
	err = json.Unmarshal(resp.Result, &messageID)

	return messageID, err
}

// AnswerWebAppQuery sets the result of an interaction with a Web App and send a
// corresponding message on behalf of the user to the chat from which the query originated.
func (bot *BotAPI) AnswerWebAppQuery(config AnswerWebAppQueryConfig) (SentWebAppMessage, error) {
	var sentWebAppMessage SentWebAppMessage

	resp, err := bot.Request(config)
	if err != nil {
		return sentWebAppMessage, err
	}

	err = json.Unmarshal(resp.Result, &sentWebAppMessage)
	return sentWebAppMessage, err
}

// GetMyDefaultAdministratorRights gets the current default administrator rights of the bot.
func (bot *BotAPI) GetMyDefaultAdministratorRights(config GetMyDefaultAdministratorRightsConfig) (ChatAdministratorRights, error) {
	var rights ChatAdministratorRights

	resp, err := bot.Request(config)
	if err != nil {
		return rights, err
	}

	err = json.Unmarshal(resp.Result, &rights)
	return rights, err
}

// EscapeText takes an input text and escape Telegram markup symbols.
// In this way we can send a text without being afraid of having to escape the characters manually.
// Note that you don't have to include the formatting style in the input text, or it will be escaped too.
// If there is an error, an empty string will be returned.
//
// parseMode is the text formatting mode (ModeMarkdown, ModeMarkdownV2 or ModeHTML)
// text is the input string that will be escaped
func EscapeText(parseMode string, text string) string {
	var replacer *strings.Replacer

	if parseMode == ModeHTML {
		replacer = strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
	} else if parseMode == ModeMarkdown {
		replacer = strings.NewReplacer("_", "\\_", "*", "\\*", "`", "\\`", "[", "\\[")
	} else if parseMode == ModeMarkdownV2 {
		replacer = strings.NewReplacer(
			"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
			"\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>",
			"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
			"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
		)
	} else {
		return ""
	}

	return replacer.Replace(text)
}

func (bot *BotAPI) SendReaction(config SetMessageReactionConfig) (Message, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return Message{}, err
	}

	var message Message
	err = json.Unmarshal(resp.Result, &message)

	return message, err
}
