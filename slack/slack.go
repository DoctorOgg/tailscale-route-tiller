package slack

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

var WebhookURL string

type SlackBlock struct {
	Type string `json:"type"`
	Text struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"text"`
}

type SlackMessage struct {
	Blocks []SlackBlock `json:"blocks"`
}

func PostRouteUpdate(subnets []string, nodeID string) {

	message := SlackMessage{
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					Type: "mrkdwn",
					Text: "*Updating Advertised routes for Node ID:* " + nodeID,
				},
			},
			{
				Type: "section",
				Text: struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					Type: "mrkdwn",
					Text: "*Subnets:*\n" + strings.Join(subnets, ", "),
				},
			},
		},
	}

	payload, err := json.Marshal(message)
	if err != nil {
		log.Fatal("Error marshaling Slack message:", err)
	}

	resp, err := http.Post(WebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal("Error sending Slack message:", err)
	}
	defer resp.Body.Close()

	log.Println("Slack message sent successfully!")
}

func PostError(err error) {

	message := SlackMessage{
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					Type: "mrkdwn",
					Text: "*Error:*\n" + err.Error(),
				},
			},
		},
	}

	payload, err := json.Marshal(message)
	if err != nil {
		log.Fatal("Error marshaling Slack message:", err)
	}

	resp, err := http.Post(WebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal("Error sending Slack message:", err)
	}
	defer resp.Body.Close()

	log.Println("Slack message sent successfully!")
}
