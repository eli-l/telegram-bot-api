package tgbotapi

type BotConfigI interface {
	GetApiEndpoint() string
	GetToken() string
	GetDebug() bool
}

type BotConfig struct {
	token string
	debug bool

	apiEndpoint string
}

func NewBotConfig(token, apiEndpoint string, debug bool) *BotConfig {
	return &BotConfig{
		token:       token,
		debug:       debug,
		apiEndpoint: apiEndpoint,
	}
}

func NewDefaultBotConfig(token string) *BotConfig {
	return &BotConfig{
		token:       token,
		debug:       false,
		apiEndpoint: APIEndpoint,
	}
}

func (c *BotConfig) GetApiEndpoint() string {
	return c.apiEndpoint
}

func (c *BotConfig) GetToken() string {
	return c.token
}

func (c *BotConfig) GetDebug() bool {
	return c.debug
}
