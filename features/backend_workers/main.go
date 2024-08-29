package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"tgs-automation/internal/opentelemetry"
	"tgs-automation/internal/util"
	"tgs-automation/pkg/telegram"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type TelegramMessage struct {
	ChatId      string `json:"chat_id"`
	Message     string `json:"message"`
	ServiceName string `json:"service_name"`
}

func main() {
	// Load configuration
	config := util.GetConfig()

	// Initialize OpenTelemetry tracer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cleanup, err := opentelemetry.InitTracer(ctx, config.OpenTelemetryConfig.OtelCollectorUrl, "domain-api", "0.1.0", "prod")

	if err != nil {
		log.Printf("Failed to init otel tracer:%v", err)
		return
	}
	defer func() {
		cancel()
		if err := cleanup(ctx); err != nil {
			log.Printf("Failed to clueanup otel tracer: %v", err)
		}
	}()

	// Initialize Telegram bot
	bot, err := initTelegramBot(config.Telegram.TelegramBotToken)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	// Connect to NATS
	nc, err := connectNATS(config.NatsUrl)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Setup JetStream
	js, err := setupJetStream(nc)
	if err != nil {
		log.Fatalf("Failed to setup JetStream: %v", err)
	}

	// Subscribe to JetStream
	sub, err := subscribeJetStream(ctx, js, "telegram.messages", bot)
	if err != nil {
		log.Fatalf("Failed to subscribe to JetStream: %v", err)
	}
	defer sub.Unsubscribe()

	// Wait for termination signal
	waitForShutdown()
}

// initTelegramBot initializes the Telegram bot instance
func initTelegramBot(token string) (*telegram.TelegramBot, error) {
	bot, err := telegram.GetTelegramBotInstance(token)
	if err != nil {
		return nil, fmt.Errorf("error getting Telegram bot instance: %w", err)
	}
	return bot, nil
}

// connectNATS connects to the NATS server
func connectNATS(natsURL string) (*nats.Conn, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to NATS: %w", err)
	}
	fmt.Println("Connected to NATS at", natsURL)
	return nc, nil
}

// setupJetStream creates the JetStream context and ensures the stream exists
func setupJetStream(nc *nats.Conn) (nats.JetStreamContext, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("error creating JetStream context: %w", err)
	}
	fmt.Println("JetStream context created")

	// Ensure the stream exists
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "TELEGRAM",
		Subjects: []string{"telegram.messages"},
	})
	if err != nil {
		return nil, fmt.Errorf("error adding stream: %w", err)
	}
	fmt.Println("JetStream stream added")

	return js, nil
}

// subscribeJetStream subscribes to a JetStream subject and handles incoming messages
func subscribeJetStream(ctx context.Context, js nats.JetStreamContext, subject string, bot *telegram.TelegramBot) (*nats.Subscription, error) {

	sub, err := js.Subscribe(subject, func(m *nats.Msg) {
		log.Printf("Received a message: %s", string(m.Data))
		if err := handleTelegramMessages(bot, m, ctx); err != nil {
			log.Printf("Error handling message: %v", err)
		}

		if err := m.Ack(); err != nil {
			log.Printf("Error acknowledging message: %v", err)
		}

	}, nats.Durable("telegram-durable"))

	if err != nil {
		return nil, fmt.Errorf("error subscribing to subject: %w", err)
	}

	fmt.Println("Subscribed to JetStream subject:", subject)
	return sub, nil
}

// waitForShutdown waits for a termination signal to gracefully shut down the application
func waitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	fmt.Println("Subscriber is running. Press Ctrl+C to stop.")

	<-stop
	fmt.Println("Shutting down...")
}

func handleTelegramMessages(bot *telegram.TelegramBot, m *nats.Msg, ctx context.Context) error {
	var msg TelegramMessage
	err := json.Unmarshal(m.Data, &msg)
	if err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return err
	}

	// headers := map[string]string{
	// 	"traceparent": m.Header.Get("traceparent"),
	// }
	propagator := propagation.TraceContext{}
	//propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	//propagator := otel.GetTextMapPropagator()

	ctx = propagator.Extract(ctx, propagation.HeaderCarrier(m.Header))

	tracer := otel.Tracer(msg.ServiceName)
	_, span := tracer.Start(ctx, "handle Telegram Messages")
	defer span.End()

	err = bot.SendMessage(msg.ChatId, msg.Message)

	if err != nil {
		log.Printf("Error sending Telegram message: %v", err)
		return err
	}

	log.Println("end handleTelegramMessages")

	return nil
}
