package bot

import (
	"pikpak-bot/utils"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TGBotListener interface {
	OnMessage(bot *TGBot, msg *tgbotapi.Message)
	OnCallbackQuery(bot *TGBot, cq *tgbotapi.CallbackQuery)
}

type TGBot struct {
	bot      *tgbotapi.BotAPI
	sessions *utils.SimpleTTLCache
	listener TGBotListener
}

func NewTGBot(token string, listener TGBotListener) (*TGBot, error) {
	bot := TGBot{
		listener: listener,
	}
	inner, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	inner.Debug = true
	bot.bot = inner
	bot.sessions = utils.NewSimpleTTLCache(time.Second * 5)
	return &bot, nil
}

func (t *TGBot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		msg := update.Message
		if msg != nil {
			t.listener.OnMessage(t, msg)
		}
		cq := update.CallbackQuery
		if cq != nil {
			t.listener.OnCallbackQuery(t, cq)
		}
	}
}

func (t *TGBot) sendMsg(chatId int64, msg string) (tgbotapi.Message, error) {
	out := tgbotapi.NewMessage(chatId, msg)
	return t.bot.Send(out)
}

func (t *TGBot) sendMsgNoResult(chatId int64, msg string) {
	out := tgbotapi.NewMessage(chatId, msg)
	_, _ = t.bot.Send(out)
}

func (t *TGBot) sendMsg2(msg tgbotapi.MessageConfig) error {
	_, err := t.bot.Send(msg)
	return err
}
