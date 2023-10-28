package bot

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"autobangumi-go/bangumi"
	"autobangumi-go/config"
	"autobangumi-go/utils"
	"github.com/pkg/errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CmdSearchBangumi    = "search_bangumi"
	CmdRestart          = "restart"
	CmdAddPikpakAccount = "add_pikpak_account"

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

type TGAutoBangumiBot struct {
	tg          *TGBot
	autoBangumi *AutoBangumi
	cache       utils.SimpleTTLCache
	cacheKey    uint64
}

func NewTGAutoBangumiBot(config *config.Config) (*TGAutoBangumiBot, error) {
	bot := TGAutoBangumiBot{}
	if config.TelegramBot.Enable {
		tgBot, err := NewTGBot(config.TelegramBot.Token, &bot)
		if err != nil {
			return nil, err
		}
		bot.tg = tgBot
		go bot.tg.Run()
	}
	autoBangumi, err := NewAutoBangumi(config)
	if err != nil {
		return nil, err
	}
	bot.autoBangumi = autoBangumi
	bot.cache = *utils.NewSimpleTTLCache(time.Hour * 24)
	autoBangumi.dl.AddCallback(&bot)
	return &bot, nil
}

func (bot *TGAutoBangumiBot) Run() {
	bot.autoBangumi.Start()
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
	case CmdAddPikpakAccount:
		return ab.executeAddPikpakAccountCmd(bot, chatId, args)
	case CmdRestart:
		bot.sendMsg("bot will be restart")
		os.Exit(0)
		return nil
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

func (ab *TGAutoBangumiBot) executeAddPikpakAccountCmd(bot *TGBot, chatId int64, args []string) error {
	if len(args) != 2 {
		return errors.Errorf("invalid add pikpak account args: %s", strings.Join(args, ","))
	}
	err := ab.autoBangumi.AddPikpakAccount(args[0], args[1])
	if err != nil {
		return errors.Wrap(err, "add pikpak account error")
	}
	bot.sendMsg("添加账号成功")
	return nil
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
			newBgm, err := bot.autoBangumi.AddBangumi(req.Name, req.TMDBId)
			if err != nil {
				tgBot.sendMsg(fmt.Sprintf("😅😅😅追番失败，错误信息： %s", err.Error()))
				return
			}
			tgBot.sendMsg("追番成功🤣, 正在后台开始下载")
			seasons, err := newBgm.GetSeasons()
			if err != nil {
				return
			}
			for _, season := range seasons {
				episodes, err := season.GetEpisodes()
				if err != nil {
					continue
				}
				tgBot.sendMsg(fmt.Sprintf("季: %d 总集数: %d 已经找到资源的集数: %d", season.GetNumber(), season.GetEpCount(), len(episodes)))
			}
		}
	}
	bot.cache.Delete(cacheKey)
}

func (bot *TGAutoBangumiBot) OnComplete(bgm bangumi.Bangumi, seasonNum uint, epNum uint) {
	bot.tg.sendMsg(fmt.Sprintf("✅ %s S%d E%d", bgm.GetTitle(), seasonNum, epNum))
}

func (bot *TGAutoBangumiBot) OnErr(err error, bgm bangumi.Bangumi, seasonNum uint, epNum uint) {
	if err == nil {
		return
	}
	bot.tg.sendMsg(fmt.Sprintf("❌ %s S%d E%d err: %s", bgm.GetTitle(), seasonNum, epNum, err.Error()))
}
