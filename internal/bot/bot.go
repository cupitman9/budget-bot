package bot

import (
	"gopkg.in/tucnak/telebot.v2"
)

func SetupHandlers(b *telebot.Bot) {
	b.Handle("/start", func(m *telebot.Message) {
		b.Send(m.Sender, "Привет")
	})
}
