package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Pashteto/tgBotTimeSqlite/config"
	"github.com/Pashteto/tgBotTimeSqlite/internal/service"
)

var ctx, _ = context.WithCancel(context.Background())

const urlTTL = time.Second * 1000

// HandlersWithDBStore  Storing data in this structure to get rid of global var DB
// data is stored using Redis DB
type HandlersWithDBStore struct {
	service FlowService

	Conf *config.Config
}

func NewHandlersWithDBStore(conf *config.Config, repo service.DbGetterSetter) *HandlersWithDBStore {
	return &HandlersWithDBStore{
		service: service.NewServ(repo),
		Conf:    conf,
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

// ==========================================================================================================================================================

// GetNikitaReq listens to bot.
func (h *HandlersWithDBStore) GetNikitaReq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Information Page</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            background-color: #000000;
            color: #ffffff;
            margin: 0;
            height: 100vh;
            display: flex;
            justify-content: flex-start;
            align-items: center;
            flex-direction: column;
            padding-left: 20px;
        }
        .content-container {
            text-align: left;
            display: flex;
            flex-direction: column;
            align-items: flex-start;
        }
        .info-block {
            transition: all 0.3s ease;
        }
        .info-block:hover {
            transform: scale(1.05);
        }
        .highlight {
            background-color: #ffffff;
            color: #000000;
            padding: 5px;
            margin: 10px 0;
            transition: all 0.3s ease;
        }
        .highlight:hover {
            background-color: #000000;
            color: #ffffff;
        }
    </style>
</head>
<body>
    <div class="content-container">
        <div class="info-block">
            <h2>Сбор на BOG:</h2>
            <p class="highlight">GE26BG0000000533615481</p>
            <p>Nikita Klimov</p>
        </div>
        <div class="info-block">
            <h2>На Тинькоф:</h2>
            <p class="highlight">+79950905198</p>
        </div>
    </div>
</body>
</html>`))

}

func (h *HandlersWithDBStore) GetTestTime(w http.ResponseWriter, r *http.Request) {
	client, err := ethclient.Dial("wss://mainnet.infura.io/ws/v3/a88b378452d94764af81d2ac741cefa7")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	contractAddress := common.HexToAddress("0x31760ad6482caf660fe48d4dfbe2953bafc31b63")
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}
	var printout []string

	for j := 0; j < 100; j++ {
		start := time.Now()
		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			log.Fatalf("Failed to filter logs: %v", err)
		}
		var printLog []string
		for _, vLog := range logs {
			// This is the hash of the Transfer event signature
			transferEventHash := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

			if vLog.Topics[0].Hex() == transferEventHash.Hex() {
				// This log is a Transfer event
				fromAddress := common.BytesToAddress(vLog.Topics[1].Bytes())
				toAddress := common.BytesToAddress(vLog.Topics[2].Bytes())
				printLog = append(printLog, fmt.Sprintf("Transfer from: %s to: %s", fromAddress.Hex(), toAddress.Hex()))
				// Now you have the "from" and "to" addresses of the Transfer
			}
		}

		endTime := time.Since(start)
		fmt.Println("Time elapsed(s): ", endTime.Seconds())
		log.Println(printLog)
		printout = append(printout, fmt.Sprintln("Time elapsed(s): ", endTime.Seconds()))

		//printout = append(printout, printLog...)
		//for _, loo := range printLog {
		//	fmt.Println(loo)
		//}
	}

	// Generate the HTML for each string
	var sb strings.Builder
	for _, str := range printout {
		sb.WriteString("<p>" + str + "</p>")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Information Page</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            background-color: #000000;
            color: #ffffff;
            margin: 0;
            height: 100vh;
            display: flex;
            justify-content: flex-start;
            align-items: center;
            flex-direction: column;
            padding-left: 20px;
        }
        .content-container {
            text-align: left;
            display: flex;
            flex-direction: column;
            align-items: flex-start;
        }
    </style>
</head>
<body>
    <div class="content-container">` + sb.String() + `</div>
</body>
</html>`))

}
