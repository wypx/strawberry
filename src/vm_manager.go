package main

import (
	"fmt"
	"log"
	HostAgent "vm_manager/host_agent"

	// VmAdmin "vm_manager/vm_admin"
	VmAdmin2 "vm_manager/vm_admin2"
	VmAgent "vm_manager/vm_agent"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TelegramBot() {
	bot, err := tgbotapi.NewBotAPI("5042048090:AAH7QLPVYLhyp1ml6d4aoAQ7smlenx2qxU0")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	fmt.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			fmt.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}

func main() {
	TelegramBot()
	log.Printf("vm manager start\n")
	// VmAdmin.Initialize()
	VmAgent.Initialize()
	HostAgent.Initialize()
	// host_agent.Initialize()
	// vm_admin.Initialize()
	VmAdmin2.Initialize()
	log.Printf("vm manager exit\n")
}
