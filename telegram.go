package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

type telegramPayload struct {
	ChatId string `json:"chat_id"`
	Text   string `json:"text"`
}

func sendTelegramMessage(message string) {
	httpClient := http.Client{}
	req := createTelegramSendMessageRequest(message)
	httpClient.Do(req)
}

func createTelegramSendMessageRequest(message string) *http.Request {
	telegramBotToken, exists := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	if !exists {
		log.Fatal("ERROR: .env variable TELEGRAM_BOT_TOKEN is missing")
	}

	telegramChatId, exists := os.LookupEnv("TELEGRAM_CHAT_ID")
	if !exists {
		log.Fatal("ERROR: .env variable TELEGRAM_CHAT_ID is missing")
	}

	payload := telegramPayload{
		ChatId: telegramChatId,
		Text:   message,
	}
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", url.PathEscape(telegramBotToken)), bytes.NewReader(encodedPayload))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("content-type", "application/json")

	return req
}
