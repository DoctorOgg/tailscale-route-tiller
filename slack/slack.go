package slack

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

var WebhookURL string
var Enabled bool = false

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

// post the message to slack
func sendit(payload []byte) {
	resp, err := http.Post(WebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal("Error sending Slack message:", err)
	}
	defer resp.Body.Close()

	log.Println("Slack message sent successfully!")
}

func PostRouteUpdate(subnets []string, nodeID string) {

	if !Enabled {
		return
	}

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

	sendit(payload)
}

func PostError(err error) {

	if !Enabled {
		return
	}

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

	sendit(payload)

}

func PostDiffUpdate(added []string, removed []string, nodeID string) {

	if !Enabled {
		return
	}

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
					Text: "*Added:*\n" + strings.Join(added, ", "),
				},
			},
			{
				Type: "section",
				Text: struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					Type: "mrkdwn",
					Text: "*Removed:*\n" + strings.Join(removed, ", "),
				},
			},
		},
	}

	payload, err := json.Marshal(message)
	if err != nil {
		log.Fatal("Error marshaling Slack message:", err)
	}

	sendit(payload)
}
