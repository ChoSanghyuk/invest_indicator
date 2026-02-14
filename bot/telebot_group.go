package bot

type TeleBotGroup struct {
	bots []*TeleBot
}

func NewTeleBotGroup(confs []*TeleBotConfig) *TeleBotGroup {

	bots := make([]*TeleBot, len(confs))
	for i, conf := range confs {
		bot, err := NewTeleBot(conf)
		if err != nil {
			panic(err) // todo. 에러 처리
		}
		bots[i] = bot
	}
	return &TeleBotGroup{
		bots: bots,
	}
}

func (t TeleBotGroup) RunAll(port int, passkey string) {
	for _, bot := range t.bots {
		go bot.Run(port, passkey)
	}
}

func (t TeleBotGroup) Bot(idx int) *TeleBot {
	if idx < 0 || idx >= len(t.bots) {
		idx = 0 // Default to the first bot if index is out of range
	}
	return t.bots[idx]
}

func (t TeleBotGroup) SendMessage(idx int, msg string) {
	if idx < 0 || idx >= len(t.bots) {
		idx = 0 // Default to the first bot if index is out of range
	}
	t.bots[idx].SendMessage(msg)
}

func (t TeleBotGroup) SendButtonsAndGetResult(idx int, prompt string, options ...string) (answer string, err error) {
	if idx < 0 || idx >= len(t.bots) {
		idx = 0 // Default to the first bot if index is out of range
	}
	return t.bots[idx].SendButtonsAndGetResult(prompt, options...)
}
