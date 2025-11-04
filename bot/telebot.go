package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TeleBot struct {
	bot     *tgbotapi.BotAPI
	chatId  int64
	updates tgbotapi.UpdatesChannel
	ch      chan string
}

type TeleBotConfig struct {
	Token  string
	ChatId int64
}

var helpMsg = `
				조회 API 목록
				/funds
				/funds/{id}/hist
				/funds/{id}/assets
				/funds/{id}/portion
				/assets
				/assets/list
				/assets/{id}
				/assets/{id}/hist
				/market
				/market/indicators/{date?}
				/events
				`

func NewTeleBot(conf *TeleBotConfig) (*TeleBot, error) {

	bot, err := tgbotapi.NewBotAPI(conf.Token) // memo. Go automatically dereferences struct pointers when accessing fields
	if err != nil {
		return nil, err
	}
	// bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	ch := make(chan string) // channel은 telebot 내부에서 thread간 소통용. telebot 외부에서는 다 동기처럼 보이게끔 설계.

	return &TeleBot{
		bot:     bot,
		chatId:  conf.ChatId,
		updates: updates,
		ch:      ch,
	}, nil
}

func (t TeleBot) Run(port int, passkey string) {
	t.SendMessage("LAUNCHED SUCCESSFULLY")

	for update := range t.updates {
		if update.Message != nil {
			txt := update.Message.Text
			if txt[0] != '/' {
				switch txt {
				case "/help":
					t.SendMessage(helpMsg)
				default:
					rtn, err := httpsend(fmt.Sprintf("http://localhost:%d%s", port, txt), passkey)
					if err != nil {
						t.SendMessage(err.Error())
					} else {
						t.SendMessage(rtn)
					}
				}
			}
		} else if update.CallbackQuery != nil {
			// Answer the callback to remove loading state
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			t.bot.Request(callback)

			// Parse and return the selected value
			t.ch <- update.CallbackQuery.Data

			newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("SUCCESSFULLY SELECTED", update.CallbackQuery.Data),
				),
			)
			// Edit the message with the updated keyboard
			editMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, newKeyboard)
			if _, err := t.bot.Send(editMsg); err != nil {
				t.SendMessage("Callback 오류. " + err.Error())
			}
		}
	}

}

func (t TeleBot) InitKey(msg error) string {

	if msg == nil {
		t.SendMessage("Enter decrypt key for invest server")
	} else {
		t.SendMessage(msg.Error())
	}
	update := <-t.updates

	return update.Message.Text
}

func (t TeleBot) SendMessage(msg string) {
	t.bot.Send(tgbotapi.NewMessage(t.chatId, msg))
}

func (t TeleBot) SendButtonsAndGetResult(prompt string, options ...string) (answer string, err error) {

	err = t.sendButtons(prompt, options...)
	if err != nil {
		return "", err
	}
	answer = <-t.ch

	return answer, nil

}

/**********************************************************************************************************************
*************************************************Inner Function*******************************************************
**********************************************************************************************************************/

// SendButtonsAndGetSelection sends a message with inline keyboard buttons for the given integer options
// and returns the selected button value. The prompt parameter is the message text to display.
func (t TeleBot) sendButtons(prompt string, options ...string) error {
	// Create inline keyboard buttons
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, option := range options {
		button := tgbotapi.NewInlineKeyboardButtonData(option, option)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Create the inline keyboard markup
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	// Create and send the message with buttons
	msg := tgbotapi.NewMessage(t.chatId, prompt)
	msg.ReplyMarkup = keyboard
	_, err := t.bot.Send(msg)
	if err != nil {
		return err
	}

	return nil

}

func httpsend(url string, passkey string) (string, error) {

	// url := "http://localhost:50001" + path
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", passkey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var jsonData interface{}

	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		return "", err
	}

	// pretty, err := json.MarshalIndent(jsonData, "", "\t") // memo. 단순 MarshalIndent 사용하면, &을 \u0026로 바꿔버림.
	// if err != nil {
	// 	return "", err
	// }

	// return string(pretty), nil
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false) // Disable HTML escaping

	// Marshal with indentation
	encoder.SetIndent("", "\t")
	err = encoder.Encode(jsonData)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

/*
var buffer bytes.Buffer
encoder := json.NewEncoder(&buffer)
encoder.SetEscapeHTML(false) // Disable HTML escaping
err := encoder.Encode(v)
return buffer.Bytes(), err
*/
