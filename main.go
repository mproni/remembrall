package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var reminders = make(map[int64]chan bool)

// Send any text message to the bot after the bot has been started
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New(os.Getenv("TELEGRAM_BOT_TOKEN_REMEMBRALL"), opts...)
	if nil != err {
		// panics for the sake of simplicity.
		// you should handle this error properly in your code.
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/hello", bot.MatchTypeExact, helloHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/set", bot.MatchTypeContains, startHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/stop", bot.MatchTypeContains, stopHandler)

	b.Start(ctx)
}

// startHandler. Mb rename it to "setHandeler()"
func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if len(update.Message.Text) < 5 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Usage: /set <message>",
		})
		return
	}
	message := update.Message.Text[len("/set "):]
	startReminder(ctx, b, update.Message.Chat.ID, message)
}

func stopHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	stopReminder(ctx, b, update.Message.Chat.ID)
}

func helloHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Hello, *" + bot.EscapeMarkdown(update.Message.From.FirstName) + "*",
		ParseMode: models.ParseModeMarkdown,
	})
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Say /set to start your reminder. Now we have only one reminder per user.",
	})
}

func startReminder(ctx context.Context, b *bot.Bot, chatID int64, message string) {
	if _, exists := reminders[chatID]; exists {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Reminder already running. Use /stop to stop it first.",
		})
		return
	}

	stop := make(chan bool)
	reminders[chatID] = stop

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   message,
				})
			case <-stop:
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   "Reminder stopped.",
				})
				return
			}
		}
	}()

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Reminder started.",
	})
}

func stopReminder(ctx context.Context, b *bot.Bot, chatID int64) {
	if stop, exists := reminders[chatID]; exists {
		stop <- true
		delete(reminders, chatID)
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "No reminder running. Use /set <message> to start one.",
		})
	}
}
