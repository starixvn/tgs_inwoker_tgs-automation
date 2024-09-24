package log

import (
	"context"
	"fmt"
	"os"
	"sync"
	"tgs-automation/internal/config"
	"tgs-automation/internal/telegram"

	log "github.com/sirupsen/logrus"
)

var (
	telegramBotInstance telegram.TelegramService
	initOnce            sync.Once
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	file, err := os.OpenFile("cdn.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to log to file, using default stderr")
	}

	log.SetOutput(file)

	ctx := context.Background()
	configObj := config.NewConfigManager()
	tgsConfig, err := configObj.GetConfig(ctx)

	if err != nil {
		fmt.Printf("Failed to get config: %v\n", err)
		log.Fatal("Failed to get config.")
		return
	}

	// 初始化 TelegramBot 實例
	initOnce.Do(func() {
		bot := telegram.NewTelegramService(tgsConfig.Telegram.TelegramBotToken, tgsConfig)
		if err != nil {
			log.Error("Failed to initialize Telegram bot:", err)
		} else {
			telegramBotInstance = bot
			log.Info("Telegram bot initialized successfully")
		}
	})
}

func LogInfo(msg string) {
	fmt.Println("Info", msg)
	log.Info(msg)
}
func LogInfoAndSendTelegram(msg string) {
	fmt.Println("Info", msg)
	log.Info(msg)

	err := telegramBotInstance.SendMessage(context.Background(), "Info: "+msg)
	if err != nil {
		fmt.Println("Failed to send Telegram message:", err)
	}
}

func LogTrace(msg string) {
	fmt.Println("Trace", msg)
	log.Trace(msg)
}

func LogError(msg string) {
	fmt.Println("Error", msg)
	log.Error(msg)
}

func LogErrorAndSendTelegram(msg string) {
	fmt.Println("Error", msg)
	log.Error(msg)
	err := telegramBotInstance.SendMessage(context.Background(), "Error "+msg)
	if err != nil {
		fmt.Println("Failed to send Telegram message:", err)
	}
}

func LogFatal(msg string) {
	fmt.Println("Fatal", msg)
	log.Fatal(msg)
}

func LogFatalAndSendTelegram(msg string) {
	fmt.Println("Fatal", msg)
	log.Fatal(msg)
	err := telegramBotInstance.SendMessage(context.Background(), "Fatal: "+msg)
	if err != nil {
		fmt.Println("Failed to send Telegram message:", err)
	}
}
