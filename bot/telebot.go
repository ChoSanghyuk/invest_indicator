package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TeleBot struct {
	bot     *tgbotapi.BotAPI
	chatId  int64
	updates tgbotapi.UpdatesChannel
}

type TeleBotConfig struct {
	Token  string
	ChatId int64
}

func NewTeleBot(conf *TeleBotConfig) (*TeleBot, error) {

	bot, err := tgbotapi.NewBotAPI(conf.Token) // memo. Go automatically dereferences struct pointers when accessing fields
	if err != nil {
		return nil, err
	}
	// bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	return &TeleBot{
		bot:     bot,
		chatId:  conf.ChatId,
		updates: updates,
	}, nil
}

func (t TeleBot) Run(ch chan string, port int, passkey string) {
	t.SendMessage("LAUNCHED SUCCESSFULLY")

	go func() {
		t.communicate(ch, port, passkey)
	}()

	for true {
		msg := <-ch
		t.SendMessage(msg)
		log.Println(msg)
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

func (t TeleBot) communicate(ch chan string, port int, passkey string) {

	for update := range t.updates {
		if update.Message != nil {
			txt := update.Message.Text
			if txt[0] != '/' {
				continue
			}

			switch txt {
			case "/help":
				ch <- `
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
			case "/form":
				ch <- `Asset
				{
				  ("id" : , )
				  "name": "",
				  "category": ,
				  "code": "",
				  "currency": "",
				  "top": ,
				  "bottom": ,
				  "ema": ,
				  "sel_price": ,
				  "buy_price": 
				}`

				ch <- `Invest
				{
				  "fund_id" : ,
				  "asset_id" : ,
				  "price" : ,
				  "count" :
				}
				`
				ch <- `AddFunds
				{
				  "name" : ""
				}
				`
				ch <- `SaveMarketStatus
				{
				  "status" : 
				}
				`
			default:
				rtn, err := httpsend(fmt.Sprintf("http://localhost:%d%s", port, txt), passkey)
				if err != nil {
					ch <- err.Error()
				} else {
					ch <- rtn
				}
			}

		}
	}
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
