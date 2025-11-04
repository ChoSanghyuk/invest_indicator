package scrape

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func TestUpbitApi(t *testing.T) {
	s := Scraper{}
	pp, cp, err := s.upbitApi("AVAX")
	if err != nil {
		t.Error(err)
	}
	t.Logf("현재가 : %f\n시가: %f", pp, cp)
}

func TestUpbitWebSocktCurrentPrice(t *testing.T) {
	// The WebSocket URL
	url := "wss://api.upbit.com/websocket/v1"

	// Create a dialer (default settings work for most cases)
	dialer := websocket.DefaultDialer

	// Optional: Set custom headers if needed
	header := http.Header{}
	// header.Add("Authorization", "Bearer YOUR_TOKEN") // if required

	// Connect to the WebSocket
	conn, resp, err := dialer.Dial(url, header)
	if err != nil {
		log.Fatalf("Failed to connect: %v, HTTP Response: %v", err, resp)
	}
	defer conn.Close()

	log.Println("Connected to WebSocket!")

	// Create your JSON message
	message := []interface{}{
		map[string]interface{}{
			"ticket": "3e2c4a9f-f0a7-457f-945e-4b57bde9f1ec", // 예제 ticket
		},
		map[string]interface{}{
			"type":  "ticker",
			"codes": []string{"KRW-BTC", "KRW-ETH"},
		},
		map[string]interface{}{
			"format": "DEFAULT",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Fatal("Marshal error:", err)
	}

	log.Printf("Sending: %s", string(jsonData))

	// Send JSON to WebSocket
	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Fatal("Write error:", err)
	}

	// Read messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		log.Printf("Received (type %d): %s", messageType, string(message))
	}
}

func TestUpbitWebSocktMyOrders(t *testing.T) {

	accessKey := os.Getenv("access")
	secretKey := os.Getenv("secret")

	payload := jwt.MapClaims{
		"access_key": accessKey,
		"nonce":      uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	jwtToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		t.Error(err)
	}

	s := Scraper{}
	s.upbit.token = jwtToken

	c := make(chan UpbitMyOrders)

	go func() {
		err = s.upbitMyOrders(c)
		fmt.Println(err)
	}()

	for true {
		msg := <-c
		fmt.Printf("%v\n", msg)
	}
}

func TestUnmarshalTest(t *testing.T) {
	var order UpbitMyOrders
	msg := []byte{123, 34, 116, 121, 112, 101, 34, 58, 34, 109, 121, 79, 114, 100, 101, 114, 34, 44, 34, 99, 111, 100, 101, 34, 58, 34, 75, 82, 87, 45, 65, 86, 65, 88, 34, 44, 34, 117, 117, 105, 100, 34, 58, 34, 102, 56, 52, 55, 56, 51, 54, 48, 45, 49, 53, 52, 56, 45, 52, 49, 50, 56, 45, 56, 97, 49, 97, 45, 53, 102, 57, 100, 53, 102, 56, 102, 101, 53, 48, 54, 34, 44, 34, 97, 115, 107, 95, 98, 105, 100, 34, 58, 34, 66, 73, 68, 34, 44, 34, 111, 114, 100, 101, 114, 95, 116, 121, 112, 101, 34, 58, 34, 108, 105, 109, 105, 116, 34, 44, 34, 115, 116, 97, 116, 101, 34, 58, 34, 119, 97, 105, 116, 34, 44, 34, 116, 114, 97, 100, 101, 95, 117, 117, 105, 100, 34, 58, 110, 117, 108, 108, 44, 34, 112, 114, 105, 99, 101, 34, 58, 50, 57, 57, 56, 48, 44, 34, 97, 118, 103, 95, 112, 114, 105, 99, 101, 34, 58, 48, 44, 34, 118, 111, 108, 117, 109, 101, 34, 58, 49, 44, 34, 114, 101, 109, 97, 105, 110, 105, 110, 103, 95, 118, 111, 108, 117, 109, 101, 34, 58, 49, 44, 34, 101, 120, 101, 99, 117, 116, 101, 100, 95, 118, 111, 108, 117, 109, 101, 34, 58, 48, 44, 34, 116, 114, 97, 100, 101, 115, 95, 99, 111, 117, 110, 116, 34, 58, 48, 44, 34, 114, 101, 115, 101, 114, 118, 101, 100, 95, 102, 101, 101, 34, 58, 49, 52, 46, 57, 57, 44, 34, 114, 101, 109, 97, 105, 110, 105, 110, 103, 95, 102, 101, 101, 34, 58, 49, 52, 46, 57, 57, 44, 34, 112, 97, 105, 100, 95, 102, 101, 101, 34, 58, 48, 44, 34, 108, 111, 99, 107, 101, 100, 34, 58, 50, 57, 57, 57, 52, 46, 57, 57, 44, 34, 101, 120, 101, 99, 117, 116, 101, 100, 95, 102, 117, 110, 100, 115, 34, 58, 48, 44, 34, 116, 105, 109, 101, 95, 105, 110, 95, 102, 111, 114, 99, 101, 34, 58, 110, 117, 108, 108, 44, 34, 116, 114, 97, 100, 101, 95, 102, 101, 101, 34, 58, 110, 117, 108, 108, 44, 34, 105, 115, 95, 109, 97, 107, 101, 114, 34, 58, 110, 117, 108, 108, 44, 34, 105, 100, 101, 110, 116, 105, 102, 105, 101, 114, 34, 58, 110, 117, 108, 108, 44, 34, 115, 109, 112, 95, 116, 121, 112, 101, 34, 58, 110, 117, 108, 108, 44, 34, 112, 114, 101, 118, 101, 110, 116, 101, 100, 95, 118, 111, 108, 117, 109, 101, 34, 58, 48, 44, 34, 112, 114, 101, 118, 101, 110, 116, 101, 100, 95, 108, 111, 99, 107, 101, 100, 34, 58, 48, 44, 34, 116, 114, 97, 100, 101, 95, 116, 105, 109, 101, 115, 116, 97, 109, 112, 34, 58, 110, 117, 108, 108, 44, 34, 111, 114, 100, 101, 114, 95, 116, 105, 109, 101, 115, 116, 97, 109, 112, 34, 58, 49, 55, 54, 49, 54, 50, 51, 50, 51, 54, 48, 48, 48, 44, 34, 116, 105, 109, 101, 115, 116, 97, 109, 112, 34, 58, 49, 55, 54, 49, 54, 50, 51, 50, 51, 54, 52, 51, 52, 44, 34, 115, 116, 114, 101, 97, 109, 95, 116, 121, 112, 101, 34, 58, 34, 82, 69, 65, 76, 84, 73, 77, 69, 34, 125}
	err := json.Unmarshal(msg, &order)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", order)
}
