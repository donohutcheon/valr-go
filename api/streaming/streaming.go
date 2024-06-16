package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/donohutcheon/valr-go/api"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	tradeWebSocketAddr   = "wss://api.valr.com/ws/trade"
	accountWebSocketAddr = "wss://api.valr.com/ws/account"

	readTimeout         = time.Minute
	writeTimeout        = 30 * time.Second
	pingInterval        = 30 * time.Second
	defaultAttemptReset = time.Minute * 30
)

type (
	ConnectCallback func(*Conn)
	UpdateCallback  func(MessageTradeUpdate)
	BackoffHandler  func(attempt int) time.Duration
)

type Conn struct {
	keyID, keySecret string
	pair             string
	connectCallback  ConnectCallback
	updateCallback   UpdateCallback

	backoffHandler BackoffHandler
	attemptReset   time.Duration

	closed bool

	mu          sync.RWMutex
	ws          *websocket.Conn
	SubscribeCh chan []string
}

// Dial initiates a connection to the streaming service and starts processing
// data for the given market pair.
// The connection will automatically reconnect on error.
func Dial(keyID, keySecret string, opts ...DialOption) (*Conn, error) {
	if keyID == "" || keySecret == "" {
		return nil, errors.New("streaming: streaming API requires credentials")
	}

	c := &Conn{
		keyID:        keyID,
		keySecret:    keySecret,
		attemptReset: defaultAttemptReset,
		SubscribeCh:  make(chan []string),
	}
	for _, opt := range opts {
		opt(c)
	}

	go c.manageForever(keyID, keySecret)
	return c, nil
}

func (c *Conn) manageForever(keyID, keySecret string) {
	p := new(backoffParams)

	for {
		if err := c.connect(keyID, keySecret); err != nil {
			log.Printf("valr/streaming: Connection error key=%s pair=%s: %v",
				c.keyID, c.pair, err)
		}
		if c.IsClosed() {
			return
		}

		dt := c.calculateBackoff(p, time.Now())

		log.Printf("valr/streaming: Waiting %s before reconnecting", dt)
		time.Sleep(dt)
	}
}

func (c *Conn) connect(keyID, keySecret string) error {
	url := tradeWebSocketAddr
	headers, err := api.GetAuthHeaders(tradeWebSocketAddr, http.MethodGet, keyID, keySecret, nil)
	if err != nil {
		return errors.Join(err, errors.New("failed to calculate auth headers"))
	}
	c.ws, _, err = websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		return fmt.Errorf("unable to dial server: %w", err)
	}
	defer func() {
		_ = c.ws.Close()
		c.reset()
	}()

	log.Printf("valr/streaming: Connection established key=%s pair=%s",
		c.keyID, c.pair)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.sendPings(ctx)

	for {
		if c.IsClosed() {
			return nil
		}

		_, data, err := c.ws.ReadMessage()
		if err != nil {
			return err
		}
		if errors.Is(err, io.EOF) {
			// Server closed the connection. Return gracefully.
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to receive message: %w", err)
		}

		if string(data) == "\"\"" {
			// Ignore server keep alive messages
			continue
		}

		fmt.Println("Received websocket payload: " + string(data))
		msgType := new(MessageType)
		err = json.Unmarshal(data, msgType)
		if err != nil {
			return fmt.Errorf("failed to unmarshal message and establish type: %w", err)
		}

		if err := c.receivedUpdate(msgType.Type, data); err != nil {
			return fmt.Errorf("failed to process update: %w", err)
		}
	}
}

func (c *Conn) receivedUpdate(msgType string, data []byte) error {
	switch msgType {
	case "NEW_TRADE":
		message := new(MessageTradeUpdate)
		err := json.Unmarshal(data, message)
		if err != nil {
			return err
		}
		fmt.Printf("%+v\n", message)
		c.updateCallback(*message)
	case "AUTHENTICATED":
		// Ignore
	case "SUBSCRIBED":
		// Ignore
	default:
		fmt.Printf("unknown message type: %s", msgType)
	}

	return nil
}

func (c *Conn) calculateBackoff(p *backoffParams, ts time.Time) time.Duration {
	if ts.Sub(p.lastAttempt) >= c.attemptReset {
		p.attempts = 0
	}

	p.attempts++

	backoff := defaultBackoffHandler
	if c.backoffHandler != nil {
		backoff = c.backoffHandler
	}

	p.lastAttempt = ts

	return backoff(p.attempts)
}

func (c *Conn) sendPings(ctx context.Context) {
	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()

	// Set initial read deadline
	_ = c.ws.SetReadDeadline(time.Now().Add(readTimeout))

	c.ws.SetPongHandler(func(data string) error {
		// Connection is alive, extend read deadline
		return c.ws.SetReadDeadline(time.Now().Add(readTimeout))
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-pingTicker.C:
			if c.IsClosed() || c.ws == nil {
				return
			}

			_ = c.ws.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("valr/streaming: Failed to ping server: %v", err)
			}
		case pairs := <-c.SubscribeCh:
			payload := SubscribeToMarketsRequest{
				Type: "SUBSCRIBE",
				Subscriptions: []Subscriptions{
					{
						Event: "NEW_TRADE",
						Pairs: pairs,
					},
				},
			}
			b, err := json.MarshalIndent(payload, "", "  ")
			if err != nil {
				log.Printf("valr/streaming: Failed to marshal payload: %v", err)
				continue
			}
			log.Printf("valr/streaming: Sending payload: %s", b)

			err = c.ws.WriteJSON(payload)
			if err != nil {
				log.Printf("valr/streaming: Failed to subscribe to pairs: %v", err)
				continue
			}
		}
	}
}

// Close the stream. After calling this the client will stop receiving new updates and the results of querying the Conn
// struct (Snapshot, Status...) will be zeroed values.
func (c *Conn) Close() {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()

	c.reset()
}

func (c *Conn) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
}

// IsClosed returns true if the Conn has been closed.
func (c *Conn) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

func (c *Conn) SubscribeToMarkets(pairs []string) {
	c.SubscribeCh <- pairs
}
