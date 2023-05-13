package bot

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"pikpak-bot/downloader"
	"pikpak-bot/utils"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	pikpakgo "github.com/lyqingye/pikpak-go"
	"github.com/rs/zerolog/log"
)

const (
	CmdDownload            = "dl"
	CmdAria2Status         = "status"
	CmdRestart             = "restart"
	CallBackTypeSelectFile = "cbsf"
)

type TGDownloaderBotConfig struct {
	Aria2WsEndpoint   string
	Aria2Secret       string
	PikpakUsername    string
	PikpakPassword    string
	DownloadDirectory string
	TGBotToken        string
}

func (cfg *TGDownloaderBotConfig) Validate() error {
	if cfg.PikpakUsername == "" || cfg.PikpakPassword == "" {
		return errors.New("pikpak username or password empty")
	}
	if cfg.TGBotToken == "" {
		return errors.New("tg bot token empty")
	}
	_, err := url.Parse(cfg.Aria2WsEndpoint)
	return err
}

type TGDownloaderBot struct {
	pikpak *pikpakgo.PikPakClient
	dl     downloader.OnlineDownloader
	cache  *utils.SimpleTTLCache
	bot    *TGBot
}

func NewTGDownloaderBot(config *TGDownloaderBotConfig) (*TGDownloaderBot, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	pikpakClient, err := pikpakgo.NewPikPakClient(config.PikpakUsername, config.PikpakPassword)
	if err != nil {
		return nil, err
	}
	err = pikpakClient.Login()
	if err != nil {
		return nil, err
	}
	aria2, err := downloader.NewAria2OnlineDownloader(config.DownloadDirectory, config.Aria2WsEndpoint, config.Aria2Secret)
	if err != nil {
		return nil, err
	}
	downloaderBot := TGDownloaderBot{
		pikpak: pikpakClient,
		dl:     aria2,
		cache:  utils.NewSimpleTTLCache(5 * time.Second),
	}
	tgBot, err := NewTGBot(config.TGBotToken, &downloaderBot)
	if err != nil {
		return nil, err
	}
	downloaderBot.bot = tgBot
	return &downloaderBot, nil
}

func (d *TGDownloaderBot) Run() {
	d.bot.Run()
}

func (d *TGDownloaderBot) OnMessage(bot *TGBot, msg *tgbotapi.Message) {
	chatId := msg.Chat.ID
	doc := msg.Document
	if doc != nil {
		if strings.HasSuffix(doc.FileName, ".torrent") {
			hash := strings.TrimSuffix(doc.FileName, ".torrent")
			hashBytes, err := hex.DecodeString(hash)
			if err != nil {
				bot.sendMsg(chatId, err.Error())
			}
			if len(hashBytes) != 20 {
				bot.sendMsg(chatId, fmt.Sprintf("invalid torrent hash: %s length: %d", hash, len(hashBytes)))
			}
			arg := fmt.Sprintf("magnet:?xt=urn:btih:%s", hex.EncodeToString(hashBytes))
			err = d.onCommand(bot, chatId, CmdDownload, []string{arg})
			if err != nil {
				bot.sendMsg(chatId, err.Error())
			}
		}
	}

	cmd := msg.Command()
	if cmd != "" {
		argsString := msg.CommandArguments()
		args := strings.Split(argsString, " ")
		err := d.onCommand(bot, chatId, cmd, args)
		if err != nil {
			bot.sendMsg(chatId, err.Error())
		}
	}
}

type FileTreeNode struct {
	Paths []string
	File  *pikpakgo.File
}

func (d *TGDownloaderBot) OnCallbackQuery(bot *TGBot, cq *tgbotapi.CallbackQuery) {
	log.Info().Str("callbackData", cq.Data).Msg("on callback query")
	chatId := cq.Message.Chat.ID
	callBackType, data := parseCallbackData(cq.Data)
	switch callBackType {
	case CallBackTypeSelectFile:
		rootFile, err := d.pikpak.GetFile(data)
		if err != nil {
			bot.sendMsgNoResult(chatId, err.Error())
			return
		}

		stack := []*FileTreeNode{
			{
				Paths: []string{},
				File:  rootFile,
			},
		}

		var flatFiles []*FileTreeNode

		for len(stack) > 0 {
			node := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if node.File.Kind == pikpakgo.KindOfFolder {
				fileList, err := d.pikpak.FileList(100, node.File.ID, "")
				if err != nil {
					bot.sendMsgNoResult(chatId, err.Error())
					return
				}

				// using for callback
				for _, f := range fileList.Files {
					stack = append(stack, &FileTreeNode{
						Paths: append(node.Paths, node.File.Name),
						File:  f,
					})
				}

			} else if node.File.Kind == pikpakgo.KindOfFile {
				flatFiles = append(flatFiles, node)
			}
		}

		for _, f := range flatFiles {
			downloadUrl, err := d.pikpak.GetDownloadUrl(f.File.ID)
			if err != nil {
				bot.sendMsgNoResult(chatId, err.Error())
				return
			}
			err = d.onCommand(bot, chatId, CmdDownload, []string{downloadUrl, f.File.Name, filepath.Join(f.Paths...)})
			if err != nil {
				bot.sendMsgNoResult(chatId, err.Error())
				return
			}
		}
	}
}

func formatCallbackData(callBackType string, data string) string {
	return fmt.Sprintf("%s|%s", callBackType, data)
}

func parseCallbackData(callbackData string) (callbackType string, data string) {
	arr := strings.Split(callbackData, "|")
	return arr[0], arr[1]
}

func (d *TGDownloaderBot) onCommand(bot *TGBot, chatId int64, cmd string, args []string) error {
	switch cmd {
	case CmdDownload:
		return d.execCmdDownload(bot, chatId, args)
	case CmdAria2Status:
		return d.execCmdAria2Status(bot, chatId)
	case CmdRestart:
		os.Exit(0)
		return nil
	default:
		return errors.New("unknown cmd")
	}
}

func (d *TGDownloaderBot) execCmdAria2Status(bot *TGBot, chatId int64) error {
	tasks, err := d.dl.ListTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		bot.sendMsg(chatId, "no tasks")
		return nil
	}
	var outputs []string
	for i, task := range tasks {
		outputs = append(outputs, fmt.Sprintf("[%d] %s %s %s %s %s", i, task.Filename, task.Status, task.Speed, task.Progress, task.ErrMessage))
	}
	_, err = bot.sendMsg(chatId, strings.Join(outputs, "\n"))
	return err
}

func (d *TGDownloaderBot) execCmdDownload(bot *TGBot, chatId int64, args []string) error {
	if len(args) == 0 || args[0] == "" {
		return errors.New("invalid dl cmd arg")
	}
	url := args[0]
	if strings.HasPrefix(url, "magnet:") {
		bot.sendMsgNoResult(chatId, "using pikpak offline download")
		newTask, err := d.pikpak.OfflineDownload("", url, "")
		if err != nil {
			return err
		}
		if newTask.Task == nil {
			return errors.New("internal error: failed to create offline task")
		}

		bot.sendMsgNoResult(chatId, fmt.Sprintf("success add pikpak offline task: %s, watting for download complete", newTask.Task.ID))
		var lastProgress int
		finishedTask, err := d.pikpak.WaitForOfflineDownloadComplete(newTask.Task.ID, time.Minute*1, func(t *pikpakgo.Task) {
			if t.Progress != lastProgress {
				bot.sendMsgNoResult(chatId, fmt.Sprintf("pikpak offline Task: %s -> %d%%", newTask.Task.ID, t.Progress))
				lastProgress = t.Progress
			}
		})
		if err != nil {
			return err
		}

		if finishedTask.Phase == pikpakgo.PhaseTypeComplete {
		} else if finishedTask.Phase == pikpakgo.PhaseTypeError {
			return errors.New(fmt.Sprintf("pikpak offline download error: %s", finishedTask.Message))
		}
		bot.sendMsgNoResult(chatId, fmt.Sprintf("pikpak offline Task: %s finished", finishedTask.ID))

		fileId := finishedTask.FileID
		file, err := d.pikpak.GetFile(fileId)
		if err != nil {
			return err
		}

		if file.Kind == pikpakgo.KindOfFolder {
			fileList, err := d.pikpak.FileList(100, fileId, "")
			if err != nil {
				return err
			}

			// using for callback
			output := tgbotapi.NewMessage(chatId, "select file to download")
			var buttons [][]tgbotapi.InlineKeyboardButton
			for _, f := range fileList.Files {
				if strings.HasSuffix(f.Name, ".torrent") {
					continue
				}
				var buttonName string
				if f.Kind == pikpakgo.KindOfFolder {
					buttonName = fmt.Sprintf("Dir: %s", f.Name)
				} else if f.Kind == pikpakgo.KindOfFile && f.Size == 0 {
					continue
				} else if f.Kind == pikpakgo.KindOfFile {
					buttonName = fmt.Sprintf("%s %s", f.Name, humanize.Bytes(uint64(f.Size)))
				}
				row := tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(buttonName, formatCallbackData(CallBackTypeSelectFile, f.ID)),
				)
				buttons = append(buttons, row)
			}
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				// download all using directory file as callback data
				tgbotapi.NewInlineKeyboardButtonData("Download all", formatCallbackData(CallBackTypeSelectFile, fileId)),
			))

			output.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
			_ = bot.sendMsg2(output)

			return nil
		} else {
			if strings.HasSuffix(file.Name, ".torrent") {
				return nil
			} else {
				downloadUrl, err := d.pikpak.GetDownloadUrl(fileId)
				if err != nil {
					return err
				}
				url = downloadUrl
				args = []string{url, file.Name}
			}
		}
	}

	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		if len(args) == 1 {
			return errors.New("you must be specify filename! example: /dl http(s):// <filename>")
		}
		filename := args[1]
		if filename == "" {
			return errors.New("filename empty")
		}
		var dir string
		if len(args) == 3 {
			dir = args[2]
		}
		taskId, err := d.dl.AddTask(url, dir, filename)
		if err != nil {
			return err
		}
		_, err = bot.sendMsg(chatId, fmt.Sprintf("add Online Download Task: %s %s", taskId, filename))
		return err
	}

	return errors.New(fmt.Sprintf("unsupported resource: %s", url))
}

type FileSelectCallBackData struct {
	FileId string
}
