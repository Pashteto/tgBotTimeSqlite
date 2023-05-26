package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/scylladb/termtables"

	"github.com/Pashteto/tgBotTimeSqlite/internal/repo"
)

type Serv struct {
	repo DbGetterSetter
}

func NewServ(repo DbGetterSetter) *Serv {
	return &Serv{repo: repo}
}

type DbGetterSetter interface {
	InsertTimezoneUserChat(tz int, name string, chat int64) error
	DelUser(chat int64, user string)
	ListChatUsers(chat int64) ([]repo.ChatMember, error)
	FindTimeZoneByChatIdAndName(name string, chat int64) (int, error)
	GetCount(chat int64) (int, error)
}

func (s *Serv) ZoneFlow(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	log.Printf("[%s] %v", update.InlineQuery.From, update.InlineQuery)
	timeShift, err := strconv.Atoi(update.InlineQuery.Query)
	if err != nil {
		s.sendHelper(bot, update.InlineQuery.ID, update.InlineQuery.Query)

		return
	}
	if timeShift > 12 || timeShift < -12 {
		s.sendHelperTime(bot, update.InlineQuery.ID, update.InlineQuery.Query, timeShift)
		return
	}
	article := tgbotapi.NewInlineQueryResultArticle(update.InlineQuery.ID, "Setting my timeZone", update.InlineQuery.Query)
	sign := " +"
	if timeShift < 0 {
		sign = " "
	}
	article.Description = fmt.Sprintf("UTC%s%d set for %s", sign, timeShift, update.InlineQuery.From.UserName)

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: update.InlineQuery.ID,
		IsPersonal:    true,
		CacheTime:     1,
		Results:       []interface{}{article},
	}
	_, err = bot.Request(inlineConf)
	if err != nil {
		log.Println(err)
		s.sendErrorInChat(bot, err.Error(), update)
	}
}

func (s *Serv) ZoneFlowMessage(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update, input string) {
	chat := update.Message.Chat
	if chat == nil || chat.ID == 0 {
		return
	}

	timeShift, err := strconv.Atoi(input)
	if len(input) == 0 {
		return
	}
	if err != nil || len(input) == 0 {
		log.Println("len(input), input, err:", len(input), input, err)
		s.sendPlease(bot, update)
		return
	}
	if timeShift > 12 || timeShift < -12 {
		s.sendErrorInChat(bot, fmt.Sprintf("your timezone %d; it should be -12 >= timezone <=12", timeShift), update)
		return
	}

	if err = s.repo.InsertTimezoneUserChat(timeShift, update.Message.From.UserName, update.Message.Chat.ID); err != nil {
		s.sendErrorInChat(bot, err.Error(), update)

		return
	}

	sign := " +"
	if timeShift < 0 {
		sign = " "
	}
	txt := fmt.Sprintf("UTC%s%d set for %s", sign, timeShift, update.Message.From.UserName)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
	bot.Send(msg)

	log.Printf("[%s] %s", update.Message.From.UserName, input)
}

func (s *Serv) SetFlow(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.InlineQuery.From.IsBot {
		s.sendHelperZone(bot, update.InlineQuery.ID, update.InlineQuery.Query, "no bots!")
		return
	}
	input := update.InlineQuery.Query

	input = strings.TrimPrefix(input, "set")
	input = strings.TrimPrefix(input, "Set")
	input = strings.TrimPrefix(input, " ")

	if len(input) == 0 {
		s.sendHelperZone(bot, update.InlineQuery.ID, update.InlineQuery.Query, "no text")
		return
	}
	input = s.addDate(input)

	tParsed, err := dateparse.ParseLocal(input)
	if err != nil {
		s.sendHelperZone(bot, update.InlineQuery.ID, update.InlineQuery.Query, err.Error())

		return
	}

	article := tgbotapi.NewInlineQueryResultArticle(update.InlineQuery.ID, "Setting meeting time:", update.InlineQuery.Query)
	article.Description = fmt.Sprintf("%s", tParsed.Format(time.TimeOnly))

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: update.InlineQuery.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{article},
	}
	_, err = bot.Request(inlineConf)
	if err != nil {
		log.Println(err)
		s.sendErrorInChat(bot, err.Error(), update)
	}
}

func (s *Serv) SetFlowMessage(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update, input string) {
	chat := update.Message.Chat
	if chat == nil || chat.ID == 0 {
		return
	}

	count, err := s.repo.GetCount(chat.ID)
	if err != nil || count == 0 {
		if err != nil {
			s.sendErrorInChat(bot, fmt.Sprintf("getCount: %s", err.Error()), update)
			return
		}

		s.sendErrorInChat(bot, fmt.Sprintf("bot has no users for this chat, add timezones!"), update)
		return
	}

	input = strings.TrimPrefix(input, "set")
	input = strings.TrimPrefix(input, "Set")
	input = strings.TrimPrefix(input, " ")

	if len(input) == 0 {
		s.sendErrorInChat(bot, fmt.Sprintf("no text"), update)
		return
	}
	input = s.addDate(input)

	tParsed, err := dateparse.ParseLocal(input)
	if err != nil {
		s.sendErrorInChat(bot, fmt.Sprintf("parse error, %s", err.Error()), update)

		return
	}

	usersShifts2 := make(map[string]int)
	usersSlice, err := s.repo.ListChatUsers(update.Message.Chat.ID)
	if err != nil {
		//sendHelperZone(bot, update.InlineQuery.ID, update.InlineQuery.Query, err.Error())
		s.sendErrorInChat(bot, fmt.Sprintf("db error, %s", err.Error()), update)

		return
	}
	for _, usr := range usersSlice {
		usersShifts2[usr.User] = usr.Shift
	}

	user := update.Message.From.UserName
	userShift, ok := usersShifts2[user]
	if !ok {
		s.sendErrorInChat(bot, fmt.Sprintf("no timezone for this user, add timezone!"), update)
		return
	}
	timeZero := s.addHours(tParsed, -userShift)

	txt := "<pre>"
	table := termtables.CreateTable()
	table.AddHeaders("Name:", "Time:", "*")
	for member, shift := range usersShifts2 {
		timeUser := s.addHours(timeZero, shift)
		org := ""
		if member == user {
			org = "*"
		}
		table.AddRow("@"+member, timeUser.Format(time.TimeOnly), org)
	}
	txt += table.Render()
	txt += "\n</pre>"
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func (s *Serv) DelFlow(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.InlineQuery.From.IsBot {
		s.sendHelperZone(bot, update.InlineQuery.ID, update.InlineQuery.Query, "no bots!")
		return
	}
	article := tgbotapi.NewInlineQueryResultArticle(update.InlineQuery.ID, "Delete my timezone data from bot", update.InlineQuery.Query)
	article.Description = fmt.Sprintf("clear %s's record from db", update.InlineQuery.From.UserName)

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: update.InlineQuery.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{article},
	}
	_, err := bot.Request(inlineConf)
	if err != nil {

		log.Println(err)
		s.sendErrorInChat(bot, err.Error(), update)
	}
}

func (s *Serv) DelFlowMessage(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	chat := update.Message.Chat
	if chat == nil || chat.ID == 0 {
		return
	}

	count, err := s.repo.GetCount(chat.ID)
	if err != nil || count == 0 {
		if err != nil {
			s.sendErrorInChat(bot, fmt.Sprintf("getCount: %s", err.Error()), update)
			return
		}

		s.sendErrorInChat(bot, fmt.Sprintf("bot has no users for this chat, add timezones!"), update)
		return
	}

	usersShifts2 := make(map[string]int)
	usersSlice, err := s.repo.ListChatUsers(update.Message.Chat.ID)
	if err != nil {
		s.sendErrorInChat(bot, fmt.Sprintf("db error, %s", err.Error()), update)

		return
	}
	for _, usr := range usersSlice {
		usersShifts2[usr.User] = usr.Shift
	}
	user := update.Message.From.UserName
	_, ok := usersShifts2[user]
	if !ok {
		txt := fmt.Sprintf("not found %s's record in db to delete", update.Message.From.UserName)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
		bot.Send(msg)

		return
	}
	s.repo.DelUser(chat.ID, update.Message.From.UserName)
	txt := fmt.Sprintf("%s's record ERASED from db", update.Message.From.UserName)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
	bot.Send(msg)
}

func (s *Serv) addHours(t time.Time, hrs int) time.Time {
	sgn := true
	if hrs < 0 {
		hrs = -hrs
		sgn = false
	}
	for i := 0; i < hrs; i++ {
		if sgn {
			t = t.Add(time.Hour)
		} else {
			t = t.Add(-time.Hour)
		}
	}
	return t
}

func (s *Serv) sendErrorInChat(bot *tgbotapi.BotAPI, er string, update *tgbotapi.Update) {
	text := fmt.Sprintf("can't set time: %s", er)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyToMessageID = update.Message.MessageID
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}
func (s *Serv) sendPlease(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please follow the examples:\n"+
		"type '+7' or '7'   <<== to set UTC+7 your timezone\n"+
		"type 'set 16:15' <<== to set your time\n"+
		"type 'del'            <<== to erase your personal data")
	msg.ReplyToMessageID = update.Message.MessageID
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (s *Serv) sendHelper(bot *tgbotapi.BotAPI, id, query string) {
	article := tgbotapi.NewInlineQueryResultArticle(id, "Helper for user/timezone definition", query)
	article.Description = "type '+/-##' example '+4' for Tbilisi"
	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: id,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{article},
	}

	if _, err := bot.Request(inlineConf); err != nil {
		log.Println(err)
	}
}
func (s *Serv) sendHelperTime(bot *tgbotapi.BotAPI, id, query string, t int) {
	article := tgbotapi.NewInlineQueryResultArticle(id, "Helper for user/timezone definition", query)
	article.Description = fmt.Sprintf("your timezone %d; it should be -12 >= timezone <=12", t)
	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: id,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{article},
	}

	if _, err := bot.Request(inlineConf); err != nil {
		log.Println(err)
	}
}
func (s *Serv) sendHelperZone(bot *tgbotapi.BotAPI, id, query string, err string) {
	article := tgbotapi.NewInlineQueryResultArticle(id, "Helper for setting meeting", query)
	article.Description = fmt.Sprintf("type something like '22:22' OR '17:00'; err == %s", err)
	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: id,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{article},
	}

	if _, err := bot.Request(inlineConf); err != nil {
		log.Println(err)
	}
}

func (s *Serv) addDate(onlyTime string) string {
	now := time.Now()
	return fmt.Sprintf("%d/%d/%d %s", now.Year(), now.Month(), now.Day(), onlyTime)
}

// generateID generates a random short ID.
func (s *Serv) generateID() (string, error) {
	var data [6]byte // 6 bytes of entropy
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data[:]), nil
}
