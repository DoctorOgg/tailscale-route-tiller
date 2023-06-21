package tailscale

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var TailScaleClientId string
var TailscaleKey string

func GetTailsScaleClientRouteSettings() ([]byte, error) {
	urlTemplate := "https://api.tailscale.com/api/v2/device/%s/routes"
	url := fmt.Sprintf(urlTemplate, TailScaleClientId)
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+TailscaleKey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	rawMessage := json.RawMessage(body)

	// Marshal the raw message with indentation
	prettyJSON, err := json.MarshalIndent(rawMessage, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return nil, err
	}

	return prettyJSON, nil
}

func SetTailscaleApprovedSubnets(subnets []string) error {
	urlTemplate := "https://api.tailscale.com/api/v2/device/%s/routes"
	url := fmt.Sprintf(urlTemplate, TailScaleClientId)

	// Create payload data
	payload := struct {
		Routes []string `json:"routes"`
	}{
		Routes: subnets,
	}

	client := &http.Client{}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error encoding JSON payload:", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+TailscaleKey)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		fmt.Println("Response:", string(body))
		return err
	}
	return nil
}
