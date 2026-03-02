package scrape

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

const (
	ProdWebSocketURL = "ws://ops.koreainvestment.com:21000"
	MockWebSocketURL = "ws://ops.koreainvestment.com:31000"
)

func (k *Kis) IssueWebSocketApprovalKey() (*WebSocketApprovalKeyResponse, error) {
	k.lg.Debug().
		Msg("IssueWebSocketApprovalKey called")

	req := &WebSocketApprovalKeyRequest{
		GrantType: "client_credentials",
		AppKey:    k.appKey,
		SecretKey: k.appSecret,
	}

	resp, err := k.executeWebSocketApprovalKey(req)
	if err != nil {
		k.lg.Error().Err(err).Msg("IssueWebSocketApprovalKey failed")
		return nil, err
	}

	k.lg.Info().
		Str("approvalKey", resp.ApprovalKey[:20]+"...").
		Msg("IssueWebSocketApprovalKey succeeded")
	return resp, nil
}

func (k *Kis) executeWebSocketApprovalKey(req *WebSocketApprovalKeyRequest) (*WebSocketApprovalKeyResponse, error) {
	endpoint := "/oauth2/Approval"
	url := k.getBaseURL() + endpoint

	k.lg.Debug().
		Str("url", url).
		Msg("Executing WebSocket approval key request")

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers (note: this endpoint does NOT require authorization token)
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	k.lg.Debug().
		Int("status", httpResp.StatusCode).
		Str("body", string(body)).
		Msg("Response received")

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", httpResp.StatusCode, string(body))
	}

	var resp WebSocketApprovalKeyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// getWebSocketURL returns the appropriate WebSocket URL based on environment
func (k *Kis) getWebSocketURL() string {
	if k.isMock {
		return MockWebSocketURL
	}
	return ProdWebSocketURL
}

// ConnectWebSocket establishes a WebSocket connection
func (k *Kis) ConnectWebSocket(approvalKey string) error {
	k.wsMutex.Lock()
	defer k.wsMutex.Unlock()

	if k.wsConn != nil {
		k.lg.Debug().Msg("WebSocket already connected")
		return nil
	}

	k.wsApprovalKey = approvalKey
	wsURL := k.getWebSocketURL()

	k.lg.Debug().
		Str("url", wsURL).
		Msg("Connecting to WebSocket")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		k.lg.Error().Err(err).Msg("Failed to connect to WebSocket")
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	k.wsConn = conn
	k.lg.Info().Msg("WebSocket connected successfully")
	return nil
}

// CloseWebSocket closes the WebSocket connection
func (k *Kis) CloseWebSocket() error {
	k.wsMutex.Lock()
	defer k.wsMutex.Unlock()

	if k.wsConn == nil {
		return nil
	}

	err := k.wsConn.Close()
	k.wsConn = nil

	if err != nil {
		k.lg.Error().Err(err).Msg("Failed to close WebSocket")
		return fmt.Errorf("failed to close WebSocket: %w", err)
	}

	k.lg.Info().Msg("WebSocket closed successfully")
	return nil
}

// SubscriptionTopic represents a subscription topic configuration
type SubscriptionTopic struct {
	TrID  string // Transaction ID (e.g., H0STCNI0 for domestic, H0GSCNI0 for overseas)
	TrKey string // HTS ID or other key for the subscription
}

// RealTimeExecutionCallbacks holds callback functions for different notification types
type RealTimeExecutionCallbacks struct {
	DomesticCallback func(*RealTimeExecutionNotification)
	OverseasCallback func(*OverseasRealTimeExecutionNotification)
}

// SubscribeMultipleRealTimeExecution subscribes to multiple real-time execution topics
// and handles responses based on the TR_ID in each message
func (k *Kis) SubscribeMultipleRealTimeExecution(
	subscribeDomestic bool,
	subscribeOverseas bool,
	callbacks *RealTimeExecutionCallbacks,
) error {
	k.wsMutex.Lock()
	if k.wsConn == nil {
		k.wsMutex.Unlock()
		return fmt.Errorf("WebSocket not connected, call ConnectWebSocket first")
	}
	k.wsMutex.Unlock()

	k.lg.Debug().
		Bool("domestic", subscribeDomestic).
		Bool("overseas", subscribeOverseas).
		Str("htsID", k.htsId).
		Msg("SubscribeMultipleRealTimeExecution called")

	// Build subscription topics list
	var topics []SubscriptionTopic

	if subscribeDomestic {
		trID := "H0STCNI0" // Production
		if k.isMock {
			trID = "H0STCNI9" // Mock
		}
		topics = append(topics, SubscriptionTopic{
			TrID:  trID,
			TrKey: k.htsId,
		})
	}

	if subscribeOverseas {
		trID := "H0GSCNI0" // Production
		if k.isMock {
			trID = "H0GSCNI9" // Mock
		}
		topics = append(topics, SubscriptionTopic{
			TrID:  trID,
			TrKey: k.htsId,
		})
	}

	if len(topics) == 0 {
		return fmt.Errorf("at least one subscription topic must be selected")
	}

	// Subscribe to each topic
	for _, topic := range topics {
		req := WebSocketSubscribeRequest{
			Header: WebSocketSubscribeRequestHeader{
				ApprovalKey: k.wsApprovalKey,
				CustType:    "P", // P for individual, B for corporate
				TrType:      "1", // 1 for subscribe, 2 for unsubscribe
				ContentType: "utf-8",
			},
			Body: WebSocketSubscribeRequestBody{
				Input: WebSocketSubscribeRequestInput{
					TrID:  topic.TrID,
					TrKey: topic.TrKey,
				},
			},
		}

		// Send subscription request
		k.wsMutex.Lock()
		err := k.wsConn.WriteJSON(req)
		k.wsMutex.Unlock()

		if err != nil {
			k.lg.Error().Err(err).Str("trID", topic.TrID).Msg("Failed to send subscription request")
			return fmt.Errorf("failed to send subscription request for %s: %w", topic.TrID, err)
		}

		k.lg.Debug().Str("trID", topic.TrID).Msg("Subscription request sent, waiting for response")

		// Read subscription response
		k.wsMutex.Lock()
		_, message, err := k.wsConn.ReadMessage()
		k.wsMutex.Unlock()

		if err != nil {
			k.lg.Error().Err(err).Str("trID", topic.TrID).Msg("Failed to read subscription response")
			return fmt.Errorf("failed to read subscription response for %s: %w", topic.TrID, err)
		}

		k.lg.Debug().
			Str("trID", topic.TrID).
			Str("message", string(message)).
			Msg("Subscription response received")

		// Parse subscription response
		var subResp WebSocketSubscribeResponse
		if err := json.Unmarshal(message, &subResp); err != nil {
			k.lg.Error().Err(err).Str("trID", topic.TrID).Msg("Failed to parse subscription response")
			return fmt.Errorf("failed to parse subscription response for %s: %w", topic.TrID, err)
		}

		// Check response
		if subResp.Body.RtCd != "0" {
			k.lg.Error().
				Str("trID", topic.TrID).
				Str("rtCd", subResp.Body.RtCd).
				Str("msg", subResp.Body.Msg1).
				Msg("Subscription failed")
			return fmt.Errorf("subscription failed for %s: code=%s, msg=%s", topic.TrID, subResp.Body.MsgCd, subResp.Body.Msg1)
		}

		// Store encryption keys per TR_ID prefix for execution notifications
		var trIDPrefix string
		if strings.HasPrefix(topic.TrID, "H0STCNI") {
			trIDPrefix = "H0STCNI" // Domestic execution notification
		} else if strings.HasPrefix(topic.TrID, "H0GSCNI") {
			trIDPrefix = "H0GSCNI" // Overseas execution notification
		}

		if trIDPrefix != "" {
			k.wsAESKeys[trIDPrefix] = subResp.Body.Output.Key
			k.wsAESIVs[trIDPrefix] = subResp.Body.Output.IV
			k.lg.Info().
				Str("trID", topic.TrID).
				Str("trIDPrefix", trIDPrefix).
				Str("key", k.wsAESKeys[trIDPrefix]).
				Str("iv", k.wsAESIVs[trIDPrefix]).
				Msg("Stored AES encryption keys for TR_ID prefix")
		}

		k.lg.Info().
			Str("trID", topic.TrID).
			Str("msg", subResp.Body.Msg1).
			Msg("Subscription successful")
	}

	// Start receiving notifications with unified handler
	k.receiveMultipleRealTimeExecutionNotifications(callbacks)

	return errors.New("receiveMultipleRealTimeExecutionNotifications exited unexpectedly")
}

// receiveMultipleRealTimeExecutionNotifications receives and routes messages based on TR_ID
func (k *Kis) receiveMultipleRealTimeExecutionNotifications(callbacks *RealTimeExecutionCallbacks) {
	for {
		k.wsMutex.Lock()
		if k.wsConn == nil {
			k.wsMutex.Unlock()
			k.lg.Debug().Msg("WebSocket connection closed, stopping notification receiver")
			return
		}
		k.wsMutex.Unlock()

		k.wsMutex.Lock()
		_, message, err := k.wsConn.ReadMessage()
		k.wsMutex.Unlock()

		if err != nil {
			k.lg.Error().Err(err).Msg("Error reading WebSocket message")
			return
		}

		messageStr := string(message)
		k.lg.Debug().
			Str("message", messageStr).
			Msg("Received WebSocket message")

		// Check for PINGPONG message and respond
		if strings.Contains(messageStr, "PINGPONG") {
			k.lg.Debug().Msg("Received PINGPONG message, sending pong")
			k.wsMutex.Lock()
			err := k.wsConn.WriteMessage(websocket.PingMessage, nil)
			k.wsMutex.Unlock()
			if err != nil {
				k.lg.Error().Err(err).Msg("Failed to send pong message")
			}
			continue
		}

		// Parse real-time message format: encrypted|TR_ID|count|data
		parts := strings.Split(messageStr, "|")
		if len(parts) < 4 {
			k.lg.Debug().
				Str("message", messageStr).
				Msg("Skipping non-realtime message")
			continue
		}

		encrypted := parts[0]
		trID := parts[1]
		// count := parts[2]
		data := parts[3]

		// Determine TR_ID prefix to get the appropriate keys
		var trIDPrefix string
		if strings.HasPrefix(trID, "H0STCNI") {
			trIDPrefix = "H0STCNI"
		} else if strings.HasPrefix(trID, "H0GSCNI") {
			trIDPrefix = "H0GSCNI"
		} else {
			k.lg.Warn().
				Str("trID", trID).
				Msg("Unknown TR_ID in real-time message")
			continue
		}

		// Get the appropriate AES keys for this TR_ID prefix
		aesKey, hasKey := k.wsAESKeys[trIDPrefix]
		aesIV, hasIV := k.wsAESIVs[trIDPrefix]

		if !hasKey || !hasIV {
			k.lg.Error().
				Str("trID", trID).
				Str("trIDPrefix", trIDPrefix).
				Bool("hasKey", hasKey).
				Bool("hasIV", hasIV).
				Msg("Missing AES keys for TR_ID prefix")
			continue
		}

		// Decrypt if encrypted
		var decryptedData string
		if encrypted == "1" {
			var err error
			decryptedData, err = decryptAES256(data, aesKey, aesIV)
			if err != nil {
				k.lg.Error().
					Err(err).
					Str("trID", trID).
					Str("trIDPrefix", trIDPrefix).
					Msg("Failed to decrypt message")
				continue
			}
		} else {
			decryptedData = data
		}

		k.lg.Debug().
			Str("trID", trID).
			Str("trIDPrefix", trIDPrefix).
			Str("decrypted", decryptedData).
			Msg("Decrypted notification data")

		// Route message based on TR_ID
		if trIDPrefix == "H0STCNI" {
			// Domestic stock execution notification
			k.handleDomesticRealTimeExecution(decryptedData, callbacks.DomesticCallback)
		} else if trIDPrefix == "H0GSCNI" {
			// Overseas stock execution notification
			k.handleOverseasRealTimeExecution(decryptedData, callbacks.OverseasCallback)
		}
	}
}

// handleDomesticRealTimeExecution parses and handles domestic execution notifications
func (k *Kis) handleDomesticRealTimeExecution(decryptedData string, callback func(*RealTimeExecutionNotification)) {
	// Parse notification (fields separated by ^)
	fields := strings.Split(decryptedData, "^")
	if len(fields) < 25 {
		k.lg.Debug().
			Int("fieldCount", len(fields)).
			Msg("Invalid domestic notification format")
		return
	}

	// Create notification object
	notification := &RealTimeExecutionNotification{
		CustID:         fields[0],
		AcctNo:         fields[1],
		OrderNo:        fields[2],
		OrigOrderNo:    fields[3],
		SellBuyDiv:     fields[4],
		ReviseDiv:      fields[5],
		OrderKind:      fields[6],
		OrderCond:      fields[7],
		StockCode:      fields[8],
		ExecQty:        fields[9],
		ExecPrice:      fields[10],
		StockExecTime:  fields[11],
		RefuseYN:       fields[12],
		ExecYN:         fields[13],
		AcceptYN:       fields[14],
		BranchNo:       fields[15],
		OrderQty:       fields[16],
		AcctName:       fields[17],
		OrderCondPrice: fields[18],
		OrderExchDiv:   fields[19],
		PopupYN:        fields[20],
		Filler:         fields[21],
		CreditDiv:      fields[22],
		CreditLoanDate: fields[23],
		ExecStockName:  fields[24],
	}
	if len(fields) > 25 {
		notification.OrderPrice = fields[25]
	}

	k.lg.Info().
		Str("orderNo", notification.OrderNo).
		Str("stockCode", notification.StockCode).
		Str("execYN", notification.ExecYN).
		Msg("Domestic real-time execution notification received")

	// Call user callback
	if callback != nil {
		callback(notification)
	}
}

// handleOverseasRealTimeExecution parses and handles overseas execution notifications
func (k *Kis) handleOverseasRealTimeExecution(decryptedData string, callback func(*OverseasRealTimeExecutionNotification)) {
	// Parse notification (fields separated by ^)
	fields := strings.Split(decryptedData, "^")
	if len(fields) < 23 {
		k.lg.Debug().
			Int("fieldCount", len(fields)).
			Msg("Invalid overseas notification format")
		return
	}

	// Create notification object
	notification := &OverseasRealTimeExecutionNotification{
		CustID:             fields[0],
		AcctNo:             fields[1],
		OrderNo:            fields[2],
		OrigOrderNo:        fields[3],
		SellBuyDiv:         fields[4],
		ReviseDiv:          fields[5],
		OrderKind2:         fields[6],
		StockShortCode:     fields[7],
		ExecQty:            fields[8],
		ExecPrice:          fields[9],
		StockExecTime:      fields[10],
		RefuseYN:           fields[11],
		ExecYN:             fields[12],
		AcceptYN:           fields[13],
		BranchNo:           fields[14],
		OrderQty:           fields[15],
		AcctName:           fields[16],
		ExecStockName:      fields[17],
		OverseasStockDiv:   fields[18],
		CollateralTypeCode: fields[19],
		CollateralLoanDate: fields[20],
		SplitBuyStartTm:    fields[21],
		SplitBuyEndTm:      fields[22],
	}
	if len(fields) > 23 {
		notification.TimeDivType = fields[23]
	}
	if len(fields) > 24 {
		notification.ExecPrice12 = fields[24]
	}

	k.lg.Info().
		Str("orderNo", notification.OrderNo).
		Str("stockCode", notification.StockShortCode).
		Str("execYN", notification.ExecYN).
		Msg("Overseas real-time execution notification received")

	// Call user callback
	if callback != nil {
		callback(notification)
	}
}

// decryptAES256 decrypts AES256 CBC encrypted data
func decryptAES256(encryptedData, key, iv string) (string, error) {
	// Decode base64 encoded data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Check IV length
	if len(iv) != aes.BlockSize {
		return "", fmt.Errorf("invalid IV size: expected %d, got %d", aes.BlockSize, len(iv))
	}

	// Decrypt using CBC mode
	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	decrypted := make([]byte, len(ciphertext))
	mode.CryptBlocks(decrypted, ciphertext)

	// Remove PKCS7 padding
	padding := int(decrypted[len(decrypted)-1])
	if padding > len(decrypted) || padding > aes.BlockSize {
		return "", fmt.Errorf("invalid padding")
	}
	decrypted = decrypted[:len(decrypted)-padding]

	return string(decrypted), nil
}
