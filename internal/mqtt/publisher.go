// Package mqtt provides the EMQX publisher used by the Go API.
package mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const defaultPublishTimeout = 5 * time.Second

// Config holds connection settings for the service MQTT publisher.
type Config struct {
	BrokerURL       string
	Username        string
	Password        string
	ClientID        string
	ConnectTimeout  time.Duration
	PublishTimeout  time.Duration
	KeepAlive       time.Duration
}

// Publisher publishes JSON events to EMQX as the privileged service account.
type Publisher struct {
	client         mqtt.Client
	publishTimeout time.Duration
	mu             sync.Mutex
}

// Connect dials EMQX and authenticates as the service account.
func Connect(cfg Config) (*Publisher, error) {
	if cfg.BrokerURL == "" {
		return nil, errors.New("mqtt broker URL is required")
	}
	if cfg.Username == "" || cfg.Password == "" {
		return nil, errors.New("mqtt service credentials are required")
	}
	if cfg.ClientID == "" {
		return nil, errors.New("mqtt client ID is required")
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 10 * time.Second
	}
	if cfg.PublishTimeout <= 0 {
		cfg.PublishTimeout = defaultPublishTimeout
	}
	if cfg.KeepAlive <= 0 {
		cfg.KeepAlive = 30 * time.Second
	}

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.BrokerURL).
		SetClientID(cfg.ClientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetKeepAlive(cfg.KeepAlive).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(2 * time.Second).
		SetOrderMatters(false).
		SetCleanSession(true)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(cfg.ConnectTimeout) {
		client.Disconnect(250)
		return nil, fmt.Errorf("mqtt connect timed out after %s", cfg.ConnectTimeout)
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("mqtt connect: %w", err)
	}

	return &Publisher{
		client:         client,
		publishTimeout: cfg.PublishTimeout,
	}, nil
}

// Publish sends raw bytes to topic at QoS 1.
func (p *Publisher) Publish(ctx context.Context, topic string, payload []byte) error {
	if p == nil || p.client == nil {
		return errors.New("mqtt publisher is not connected")
	}
	if topic == "" {
		return errors.New("mqtt topic is required")
	}

	timeout := p.publishTimeout
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > 0 && remaining < timeout {
			timeout = remaining
		}
	}

	p.mu.Lock()
	token := p.client.Publish(topic, 1, false, payload)
	p.mu.Unlock()

	if !token.WaitTimeout(timeout) {
		return fmt.Errorf("mqtt publish to %s timed out", topic)
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("mqtt publish to %s: %w", topic, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

// PublishEvent marshals event as JSON and publishes it to topic.
func (p *Publisher) PublishEvent(ctx context.Context, topic string, event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal mqtt event: %w", err)
	}
	return p.Publish(ctx, topic, body)
}

// PublishToUserInbox publishes event to chat/user/{userID}/inbox.
func (p *Publisher) PublishToUserInbox(ctx context.Context, userID string, event Event) error {
	if userID == "" {
		return errors.New("user id is required")
	}
	return p.PublishEvent(ctx, UserInboxTopic(userID), event)
}

// Close disconnects from EMQX.
func (p *Publisher) Close() {
	if p == nil || p.client == nil {
		return
	}
	p.client.Disconnect(250)
}
