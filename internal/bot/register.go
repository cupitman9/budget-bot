package bot

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/tucnak/telebot.v2"

	"budget-bot/internal/model"
	"budget-bot/internal/storage"
)

func RegisterHandlers(b *telebot.Bot, storageInstance *storage.Storage, log *logrus.Logger, userSessions map[int64]*model.UserSession) {
	cbHandler := newCallbackHandler(b, storageInstance, log)
	msgHandler := newMessageHandler(b, storageInstance, log)

	b.Handle("/start", func(m *telebot.Message) {
		log.Infof("processing /start for user %d", m.Sender.ID)
		msgHandler.handleStart(m)
	})

	b.Handle("/help", func(m *telebot.Message) {
		log.Infof("processing /help for user %d", m.Sender.ID)
		msgHandler.handleHelp(m)
	})

	b.Handle("/add_category", func(m *telebot.Message) {
		log.Infof("processing /add_category for user %d", m.Sender.ID)
		userSessions[m.Sender.ID] = &model.UserSession{
			State: model.StateAwaitingNewCategoryName,
		}
		b.Send(m.Sender, "Введите название новой категории:")
	})

	b.Handle("/show_categories", func(m *telebot.Message) {
		log.Infof("processing /show_categories for user %d", m.Sender.ID)
		msgHandler.handleShowCategories(m)
	})

	b.Handle("/stats", func(m *telebot.Message) {
		log.Infof("processing /stats for user %d", m.Sender.ID)
		msgHandler.handleStatsButtons(m)
	})

	b.Handle(telebot.OnText, func(m *telebot.Message) {
		log.Infof("processing text from user %d: %s", m.Sender.ID, m.Text)
		msgHandler.handleOnText(m, userSessions)
	})

	b.Handle(telebot.OnCallback, func(c *telebot.Callback) {
		log.Infof("processing callback from user %d: %s", c.Sender.ID, c.Data)
		cbHandler.handleCallback(c, userSessions)
	})
}
