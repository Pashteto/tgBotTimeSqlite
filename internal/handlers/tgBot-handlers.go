package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/websocket"

	"github.com/Pashteto/tgBotTimeSqlite/config"
	"github.com/Pashteto/tgBotTimeSqlite/internal/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ctx, _ = context.WithCancel(context.Background())

const urlTTL = time.Second * 1000

// HandlersWithDBStore  Storing data in this structure to get rid of global var DB
// data is stored using Redis DB
type HandlersWithDBStore struct {
	service FlowService

	Conf *config.Config

	FileServer *http.Handler
}

func NewHandlersWithDBStore(conf *config.Config, repo service.DbGetterSetter, fs *http.Handler) *HandlersWithDBStore {
	return &HandlersWithDBStore{
		service:    service.NewServ(repo),
		Conf:       conf,
		FileServer: fs,
	}
}

type FlowService interface {
	ZoneFlow(bot *tgbotapi.BotAPI, update *tgbotapi.Update)
	ZoneFlowMessage(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update, input string)
	SetFlow(bot *tgbotapi.BotAPI, update *tgbotapi.Update)
	SetFlowMessage(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update, input string)
	DelFlow(bot *tgbotapi.BotAPI, update *tgbotapi.Update)
	DelFlowMessage(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update)
}

func (h *HandlersWithDBStore) EmptyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

var stopServiceListen chan struct{}

// ListenBot listens to bot.
func (h *HandlersWithDBStore) ListenBot(w http.ResponseWriter, r *http.Request) {
	if stopServiceListen != nil {
		log.Println("the service is running, no need to start another listener")
		http.Error(w, "the service is running, no need to start another listener", http.StatusBadRequest)
		return
	}

	stopServiceListen = make(chan struct{})
	go func() {
		err := func() error {
			err := h.startListener(ctx, stopServiceListen)
			if err != nil {

				return err
			}

			return nil
		}()
		if err != nil {
			log.Printf("ERROR FROM startListener, %s", err.Error())
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	write, err := w.Write([]byte("the service is running"))
	if err != nil {
		log.Printf("cant write resp: % bytes, err: %s", write, err.Error())
	}

	return
}

// StopListenBot listens to bot.
func (h *HandlersWithDBStore) StopListenBot(w http.ResponseWriter, r *http.Request) {
	if stopServiceListen == nil {
		log.Println("the service is not running, no need to stop it")
		http.Error(w, "the service is not running, no need to stop it", http.StatusBadRequest)

		return
	}

	stopServiceListen <- struct{}{}
	time.Sleep(2 * time.Second)
	stopServiceListen = nil

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	write, err := w.Write([]byte("the service is stopped"))
	if err != nil {
		log.Printf("cant write resp: % bytes, err: %s", write, err.Error())
	}

	return
}

// startListener starts to listen the bot.
func (h *HandlersWithDBStore) startListener(ctx context.Context, stop chan struct{}) error {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Println("TELEGRAM_TOKEN = ", os.Getenv("TELEGRAM_TOKEN"), "error: ", err.Error())
		log.Panic(err.Error())
	}
	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 3600

	updates := bot.GetUpdatesChan(u)

	//chatsUserTimeShifts := make(map[int64]map[string]int)
	nonStop := true
	botUserName := "@" + bot.Self.UserName

	for nonStop {
		select {
		case <-stop:
			log.Printf("stop signal received")
			nonStop = false
			bot.StopReceivingUpdates()
			break
		case update := <-updates:
			inlineFlow := true
			if inlineFlow {
				if update.InlineQuery != nil {
					if strings.HasPrefix(update.InlineQuery.Query, "Del") ||
						strings.HasPrefix(update.InlineQuery.Query, "del") {
						h.service.DelFlow(bot, &update)

						continue
					}

					if !strings.HasPrefix(update.InlineQuery.Query, "set") &&
						!strings.HasPrefix(update.InlineQuery.Query, "Set") {
						h.service.ZoneFlow(bot, &update)

						continue
					}
					h.service.SetFlow(bot, &update)

					continue
				}
			}

			if update.Message != nil { // If we got a message
				input := update.Message.Text

				mentionFlow := false
				if mentionFlow {

					mention := false
					replyToMe := false
					for _, entity := range update.Message.Entities {
						if entity.Type == "mention" && update.Message.From != nil && update.Message.From.UserName != bot.Self.UserName {
							mention = true
							break
						}
					}

					if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.UserName == bot.Self.UserName {
						replyToMe = true
					}

					if !mention || !strings.Contains(update.Message.Text, botUserName) {
						if !replyToMe {
							continue
						}
					}
					txt := update.Message.Text
					input = strings.TrimRight(strings.TrimLeft(strings.Join(strings.Split(txt, botUserName), ""), " !?"), " !?")

					// The message mentions the bot, so do something with it
					log.Printf("The message mentions the bot:%s", update.Message.Text)
					log.Printf("The message wo mention:%s", input)
					log.Printf("botUserName:%s", botUserName)

				}

				if strings.HasPrefix(input, "Del") ||
					strings.HasPrefix(input, "del") {
					h.service.DelFlowMessage(ctx, bot, &update)

					continue
				}
				if !strings.HasPrefix(input, "set") &&
					!strings.HasPrefix(input, "Set") {
					h.service.ZoneFlowMessage(ctx, bot, &update, input)

					continue
				}
				h.service.SetFlowMessage(ctx, bot, &update, input)

				continue
			}
		}
	}
	//for update := range updates {	}

	return nil
}
