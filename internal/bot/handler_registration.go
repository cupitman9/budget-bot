package bot

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"

	"github.com/cupitman9/budget-bot/internal/model"
	"github.com/cupitman9/budget-bot/internal/storage"
)

var userSessions = make(map[int64]*model.UserSession)

func RegisterHandlers(b *telebot.Bot, storageInstance *storage.Storage, log *logrus.Logger) {
	cbHandler := newCallbackHandler(b, storageInstance, log)
	msgHandler := newMessageHandler(b, storageInstance, log)

	b.Handle("/start", func(ctx telebot.Context) error {
		err := msgHandler.handleStart(ctx.Message())
		if err != nil {
			log.WithField("userId", ctx.Message().Sender.ID).WithError(err).Error("error handling /start")
		}
		return nil
	})

	b.Handle("/help", func(ctx telebot.Context) error {
		err := msgHandler.handleHelp(ctx.Message())
		if err != nil {
			log.WithField("userId", ctx.Message().Sender.ID).WithError(err).Error("error handling /help")
		}
		return nil
	})

	b.Handle("/add_category", func(ctx telebot.Context) error {
		userSessions[ctx.Message().Sender.ID] = &model.UserSession{
			State: model.StateAwaitingNewCategoryName,
		}

		_, err := b.Send(ctx.Sender(), "Введите название новой категории:")
		if err != nil {
			log.WithField("userId", ctx.Message().Sender.ID).WithError(err).
				Error("error handling /add_category")
		}
		return nil
	})

	b.Handle("/show_categories", func(ctx telebot.Context) error {
		err := msgHandler.handleShowCategories(ctx.Message())
		if err != nil {
			log.WithField("userId", ctx.Message().Sender.ID).WithError(err).
				Error("error handling /show_categories")
		}
		return nil
	})

	b.Handle("/stats", func(ctx telebot.Context) error {
		err := msgHandler.handleStatsButtons(ctx.Message())
		if err != nil {
			log.WithField("userId", ctx.Message().Sender.ID).WithError(err).Error("error handling /stats")
		}
		return nil
	})

	b.Handle(telebot.OnText, func(ctx telebot.Context) error {
		err := msgHandler.handleOnText(ctx.Message())
		if err != nil {
			log.WithField("userId", ctx.Message().Sender.ID).WithError(err).Error("error handling text")
		}
		return nil
	})

	b.Handle(telebot.OnCallback, func(ctx telebot.Context) error {
		err := cbHandler.handleCallback(ctx.Callback())
		if err != nil {
			log.WithField("userId", ctx.Message().Sender.ID).WithError(err).Error("error handling callback")
		}
		return nil
	})
}
