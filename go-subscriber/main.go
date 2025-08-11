package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

// Message represents the structure of the incoming message.
type Message struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Sender    string `json:"sender"`
	Count     int    `json:"count"`
}

// Subscriber encapsulates logic and state for the Redis subscriber.
type Subscriber struct {
	client       *redis.Client
	ctx          context.Context
	cancel       context.CancelFunc
	pattern      string
	messageCount int
	mu           sync.Mutex
}

// NewSubscriber initializes a Redis client and returns a Subscriber instance.
func NewSubscriber() *Subscriber {
	ctx, cancel := context.WithCancel(context.Background())

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	pattern := getEnv("PATTERN", "both") // Options: 'pubsub', 'queue', or 'both'

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	return &Subscriber{
		client:  client,
		ctx:     ctx,
		cancel:  cancel,
		pattern: pattern,
	}
}

// getEnv returns the value of an environment variable, or a default if not set.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// connect tests the Redis connection.
func (s *Subscriber) connect() error {
	_, err := s.client.Ping(s.ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Printf("[%s] Connected to Redis at %s", time.Now().Format(time.RFC3339), s.client.Options().Addr)
	return nil
}

// parseAndDisplayMessage parses JSON and prints a structured log message.
func (s *Subscriber) parseAndDisplayMessage(messageData string, source string) {
	var msg Message
	if err := json.Unmarshal([]byte(messageData), &msg); err != nil {
		log.Printf("[%s] Failed to parse message from %s: %v", time.Now().Format(time.RFC3339), source, err)
		return
	}

	s.mu.Lock()
	s.messageCount++
	count := s.messageCount
	s.mu.Unlock()

	timestamp, err := time.Parse(time.RFC3339, msg.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	fmt.Printf("[%s] [%s] %s (ID: %s) [Received #%d from %s]\n",
		timestamp.Format("2006-01-02 15:04:05"),
		msg.Sender,
		msg.Message,
		msg.ID,
		count,
		source)
}

// subscribePubSub subscribes to a Redis pub-sub channel and listens for messages.
func (s *Subscriber) subscribePubSub() {
	log.Printf("[%s] Starting pub-sub subscriber on channel 'messages'", time.Now().Format(time.RFC3339))

	pubsub := s.client.Subscribe(s.ctx, "messages")
	defer pubsub.Close()

	if _, err := pubsub.Receive(s.ctx); err != nil {
		log.Printf("[%s] Failed to confirm subscription: %v", time.Now().Format(time.RFC3339), err)
		return
	}

	log.Printf("[%s] Successfully subscribed to 'messages' channel", time.Now().Format(time.RFC3339))

	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg != nil {
				s.parseAndDisplayMessage(msg.Payload, "pub-sub")
			}
		case <-s.ctx.Done():
			log.Printf("[%s] Pub-sub subscriber shutting down", time.Now().Format(time.RFC3339))
			return
		}
	}
}

// subscribeQueue uses BLPOP to consume messages from a Redis list (queue).
func (s *Subscriber) subscribeQueue() {
	log.Printf("[%s] Starting queue subscriber on list 'message_queue'", time.Now().Format(time.RFC3339))

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("[%s] Queue subscriber shutting down", time.Now().Format(time.RFC3339))
			return
		default:
			result, err := s.client.BLPop(s.ctx, 1*time.Second, "message_queue").Result()
			if err != nil {
				if err == redis.Nil || err == context.Canceled {
					continue
				}
				log.Printf("[%s] Error in BLPOP: %v", time.Now().Format(time.RFC3339), err)
				time.Sleep(1 * time.Second)
				continue
			}

			if len(result) == 2 {
				s.parseAndDisplayMessage(result[1], "queue")
			}
		}
	}
}

// run starts one or both message listeners and handles graceful shutdown.
func (s *Subscriber) run() error {
	log.Printf("[%s] Starting Go Subscriber...", time.Now().Format(time.RFC3339))
	log.Printf("[%s] Pattern: %s", time.Now().Format(time.RFC3339), s.pattern)

	// Retry connection attempts
	maxRetries := 5
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := s.connect(); err != nil {
			log.Printf("[%s] Connection attempt %d/%d failed: %v",
				time.Now().Format(time.RFC3339), attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(5 * time.Second)
				continue
			}
			return fmt.Errorf("failed to connect after %d attempts", maxRetries)
		}
		break
	}

	var wg sync.WaitGroup

	if s.pattern == "pubsub" || s.pattern == "both" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.subscribePubSub()
		}()
	}

	if s.pattern == "queue" || s.pattern == "both" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.subscribeQueue()
		}()
	}

	// Handle shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Printf("[%s] Received shutdown signal. Shutting down gracefully...", time.Now().Format(time.RFC3339))
	s.cancel()
	wg.Wait()

	log.Printf("[%s] Subscriber stopped. Total messages received: %d", time.Now().Format(time.RFC3339), s.messageCount)

	if err := s.client.Close(); err != nil {
		log.Printf("[%s] Error closing Redis client: %v", time.Now().Format(time.RFC3339), err)
	}

	return nil
}

// main is the entry point of the application.
func main() {
	subscriber := NewSubscriber()
	if err := subscriber.run(); err != nil {
		log.Fatalf("Subscriber failed: %v", err)
	}
}
