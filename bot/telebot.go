package bot

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TeleBot struct {
	bot     *tgbotapi.BotAPI
	chatId  int64
	updates tgbotapi.UpdatesChannel
}

func NewTeleBot(token string, chatId int64) (*TeleBot, error) {

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	return &TeleBot{
		bot:     bot,
		chatId:  chatId,
		updates: updates,
	}, nil
}

func (t TeleBot) InitKey() string {
	t.SendMessage("Enter decrypt key for invest server")
	update := <-t.updates

	return update.Message.Text
}

func (t TeleBot) SendMessage(msg string) {
	t.bot.Send(tgbotapi.NewMessage(t.chatId, msg))
}

func (t TeleBot) Listen(ch chan string) {

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
				/funds/:{id}/assets
				/assets/list
				/assets/{id}
				/assets/{id}/hist
				/market
				/market/indicators/{date?}
				`
			case "/form":
				ch <- `
				Asset
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
				}

				Invest
				{
				  "fund_id" : ,
				  "asset_id" : ,
				  "price" : ,
				  "count" :
				}

				AddFunds
				{
				  "name" : ""
				}
				
				SaveMarketStatus
				{
				  "status" : 
				}
				`
			default:
				rtn, err := httpsend(txt)
				if err != nil {
					ch <- err.Error()
				} else {
					ch <- rtn
				}
			}

		}
	}
}

func httpsend(path string) (string, error) {

	url := "http://localhost:3000" + path
	req, _ := http.NewRequest(http.MethodGet, url, nil)

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
