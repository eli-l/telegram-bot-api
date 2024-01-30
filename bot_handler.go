package tgbotapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type HandlerConfig struct {
	shutdownChannel chan struct{}
	bufferSize      int
}

var defaultConfig = HandlerConfig{
	shutdownChannel: make(chan struct{}),
	bufferSize:      100,
}

type PollingHandler struct {
	bot *BotAPI
	HandlerConfig
	updateConfig UpdateConfig
}

func NewPollingHandler(bot *BotAPI, updateConfig UpdateConfig) *PollingHandler {
	return &PollingHandler{
		bot:           bot,
		HandlerConfig: defaultConfig,
		updateConfig:  updateConfig,
	}
}

type WebhookHandler struct {
	bot *BotAPI
	HandlerConfig
}

func NewWebhookHandler(bot *BotAPI) *WebhookHandler {
	return &WebhookHandler{
		bot:           bot,
		HandlerConfig: defaultConfig,
	}
}

// InitUpdatesChannel starts and returns a channel for getting updates.
func (h *PollingHandler) InitUpdatesChannel() (UpdatesChannel, error) {
	w, err := h.bot.GetWebhookInfo()
	if err == nil && w.IsSet() {
		return nil, errors.New("webhook was set, can't use polling")
	}

	ch := make(chan Update, h.bufferSize)

	go func() {
		for {
			select {
			case <-h.shutdownChannel:
				close(ch)
				return
			default:
			}

			updates, err := h.bot.GetUpdates(h.updateConfig)
			if err != nil {
				log.Println(err)
				log.Println("Failed to get updates, retrying in 3 seconds...")
				time.Sleep(time.Second * 3)

				continue
			}

			for _, update := range updates {
				if update.UpdateID >= h.updateConfig.Offset {
					h.updateConfig.Offset = update.UpdateID + 1
					ch <- update
				}
			}
		}
	}()

	return ch, nil
}

// Stop stops the go routine which receives updates
func (h *PollingHandler) Stop() {
	if h.bot.GetConfig().GetDebug() {
		log.Println("Stopping the update receiver routine...")
	}
	close(h.shutdownChannel)
}

// ListenForWebhook registers a http handler for a webhook.
func (h *WebhookHandler) ListenForWebhook(pattern string) UpdatesChannel {
	ch := make(chan Update, h.bufferSize)

	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		update, err := UnmarshalUpdate(r)
		if err != nil {
			errMsg, _ := json.Marshal(map[string]string{"error": err.Error()})
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(errMsg)
			return
		}

		ch <- *update
	})

	return ch
}

// ListenForWebhookRespReqFormat registers a http handler for a single incoming webhook.
func (h *WebhookHandler) ListenForWebhookRespReqFormat(w http.ResponseWriter, r *http.Request) UpdatesChannel {
	ch := make(chan Update, h.bufferSize)

	func(w http.ResponseWriter, r *http.Request) {
		defer close(ch)

		update, err := UnmarshalUpdate(r)
		if err != nil {
			errMsg, _ := json.Marshal(map[string]string{"error": err.Error()})
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(errMsg)
			return
		}

		ch <- *update
	}(w, r)

	return ch
}

// UnmarshalUpdate parses and returns update received via webhook
func UnmarshalUpdate(r *http.Request) (*Update, error) {
	if r.Method != http.MethodPost {
		err := errors.New("wrong HTTP method required POST")
		return nil, err
	}

	var update Update
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		return nil, err
	}

	return &update, nil
}
