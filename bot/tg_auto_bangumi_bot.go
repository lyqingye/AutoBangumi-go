package bot

import (
	"autobangumi-go/bangumi"
	"autobangumi-go/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CmdSearchBangumi = "search_bangumi"

	CallBackTypeAddNewBangumi = "cb_add_bangumi"
)

type AddNewBangumiRequest struct {
	Name   string
	TMDBId int64
}

func formatCallbackData(callBackType string, cacheKey uint64) string {
	return fmt.Sprintf("%s|%d", callBackType, cacheKey)
}

func parseCallbackData(callbackData string) (callbackType string, cacheKey uint64, err error) {
	arr := strings.Split(callbackData, "|")
	cacheKey, err = strconv.ParseUint(arr[1], 10, 64)
	if err != nil {
		return
	}
	return arr[0], cacheKey, nil
}

type TGAutoBangumiBotConfig struct {
	AutoBangumiConfig
	TGBotToken string
}

type TGAutoBangumiBot struct {
	tg          *TGBot
	autoBangumi *AutoBangumi
	cache       utils.SimpleTTLCache
	cacheKey    uint64
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
	bot.cache = *utils.NewSimpleTTLCache(time.Hour * 24)
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
			bot.sendMsgNoResult(err.Error())
		}
	}
}

func (ab *TGAutoBangumiBot) onCommand(bot *TGBot, chatId int64, cmd string, args []string) error {
	switch cmd {
	case CmdSearchBangumi:
		return ab.executeSearchBangumiCmd(bot, chatId, args)
	default:
		return fmt.Errorf("unknown cmd: %s", cmd)
	}
}

func (ab *TGAutoBangumiBot) executeSearchBangumiCmd(bot *TGBot, chatId int64, args []string) error {
	tmdb := ab.autoBangumi.tmdb
	if len(args) != 1 {
		return errors.New("invalid args")
	}
	keyword := args[0]
	tv, err := tmdb.SearchTVShowByKeyword(keyword)
	if err != nil {
		return err
	}
	picUrl := fmt.Sprintf("https://www.themoviedb.org/t/p/w440_and_h660_bestv2/%s", tv.PosterPath)
	resp, err := http.Get(picUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	msg := tgbotapi.NewPhoto(chatId, tgbotapi.FileReader{
		Name:   tv.Name,
		Reader: resp.Body,
	})
	msg.Caption = tv.Name
	cacheKey := atomic.AddUint64(&ab.cacheKey, 1)
	callbackData := formatCallbackData(CallBackTypeAddNewBangumi, cacheKey)
	ab.cache.Put(cacheKey, AddNewBangumiRequest{
		Name:   tv.Name,
		TMDBId: tv.ID,
	}, time.Hour*24)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("开始追番😅", callbackData)})
	_, err = bot.bot.Send(msg)
	return err
}

func (bot *TGAutoBangumiBot) OnCallbackQuery(tgBot *TGBot, cq *tgbotapi.CallbackQuery) {
	cqType, cacheKey, err := parseCallbackData(cq.Data)
	if err != nil {
		bot.tg.sendMsg(err.Error())
		return
	}
	switch cqType {
	case CallBackTypeAddNewBangumi:
		if value, found := bot.cache.Get(cacheKey); found {
			req := value.(AddNewBangumiRequest)
			bgmMan := bot.autoBangumi.bgmMan
			if bgmMan.IsBangumiExist(req.Name) {
				tgBot.sendMsg("已经追番成功，请勿重复追番😅😅😅")
			} else {
				bgmMan.AddBangumiIfNotExist(bangumi.Bangumi{
					Info: bangumi.BangumiInfo{
						Title:  req.Name,
						TmDBId: req.TMDBId,
					},
				})
				tgBot.sendMsg("追番成功🤣")
			}
		}
	}
	bot.cache.Delete(cacheKey)
}
