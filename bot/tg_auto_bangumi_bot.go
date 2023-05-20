package bot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CmdAddMikanSubcribeLink = "rss_sub"
	CmdRefreshSubcribe      = "rss_refresh"
)

type TGAutoBangumiBotConfig struct {
	AutoBangumiConfig
	TGBotToken string
}

type TGAutoBangumiBot struct {
	tg          *TGBot
	autoBangumi *AutoBangumi
}

func NewTGAutoBangumiBot(config *TGAutoBangumiBotConfig) (*TGAutoBangumiBot, error) {
	bot := TGAutoBangumiBot{}
	tgBot, err := NewTGBot(config.TGBotToken, &bot)
	if err != nil {
		return nil, err
	}
	bot.tg = tgBot
	autoBangumi, err := NewAutoBangumi(&config.AutoBangumiConfig)
	if err != nil {
		return nil, err
	}
	bot.autoBangumi = autoBangumi
	return &bot, nil
}

func (bot *TGAutoBangumiBot) Run() {
	go bot.autoBangumi.Start()
	bot.tg.Run()
}

func (ab *TGAutoBangumiBot) OnMessage(tgBot *TGBot, msg *tgbotapi.Message) {
	bot := ab.tg
	chatId := msg.Chat.ID
	cmd := msg.Command()
	if cmd != "" {
		argsString := msg.CommandArguments()
		args := strings.Split(argsString, " ")
		err := ab.onCommand(bot, chatId, cmd, args)
		if err != nil {
			bot.sendMsg(chatId, err.Error())
		}
	}
}

func (ab *TGAutoBangumiBot) onCommand(bot *TGBot, chatId int64, cmd string, args []string) error {
	switch cmd {
	case CmdAddMikanSubcribeLink:
		return ab.execCmdDownload(bot, chatId, args)
	case CmdRefreshSubcribe:
		go ab.autoBangumi.rssMan.Refresh()
		return nil
	default:
		return fmt.Errorf("unknown cmd: %s", cmd)
	}
}

func (ab *TGAutoBangumiBot) execCmdDownload(bot *TGBot, chatId int64, args []string) error {
	if len(args) != 1 {
		return errors.New("invalid arguments")
	}
	arg := args[0]
	bangumiId, err := strconv.ParseInt(arg, 10, 64)
	if err == nil {
		arg = fmt.Sprintf("https://mikanani.me/RSS/Bangumi?bangumiId=%d", bangumiId)
	}

	if strings.HasPrefix(arg, "https://mikanani.me/RSS") {
		// err := ab.autoBangumi.AddMikanRss(arg)
		// if err != nil {
		// 	return err
		// }
		bot.sendMsg(chatId, fmt.Sprintf("success subscribe mikan rss! link: %s", arg))
		return nil
	}

	return errors.New("unknown resource")
}

func (bot *TGAutoBangumiBot) OnCallbackQuery(tgBot *TGBot, cq *tgbotapi.CallbackQuery) {

}
