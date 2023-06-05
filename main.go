package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"tailscale-route-tiller/slack"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var subnets []string

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type Slack struct {
	WebhookURL string `yaml:"WebhookURL"`
	Enabled    bool   `yaml:"Enabled"`
}

// Config is a struct for our YAML data
type Config struct {
	Subnets           []string `yaml:"subnets"`
	Sites             []string `yaml:"sites"`
	TailscaleCommand  string   `yaml:"TailscaleCommand"`
	EnableIpv6        bool     `yaml:"EnableIpv6"`
	TailscaleclientId string   `yaml:"TailscaleclientId"`
	TailscaleKey      string   `yaml:"TailscaleKey"`
	Slack             Slack    `yaml:"Slack"`
}

type IPSubnet struct {
	Network net.IPNet
}

var ActiveConfig *Config

// ReadYAML reads the YAML configuration file
func ReadYAML(filename string) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading YAML file: ", err)
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}
		os.Exit(1)
	}

	c := &Config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		fmt.Println("Error reading YAML file: ", err)
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}
		os.Exit(1)
	}
	ActiveConfig = c
}

func PerformDNSLookups(sites []string) []string {
	var subnetsList []string
	for _, site := range sites {
		ips, err := net.LookupIP(site)
		if err != nil {
			fmt.Println("Error: ", site, err.Error())
			if ActiveConfig.Slack.Enabled {
				slack.WebhookURL = ActiveConfig.Slack.WebhookURL
				slack.PostError(err)
			}

		} else {
			for _, ip := range ips {
				mask := "/32"
				if ip.To4() == nil { // it's IPv6
					mask = "/128"
					if ActiveConfig.EnableIpv6 {
						subnetsList = append(subnetsList, ip.String()+mask)
						fmt.Println("IP addresses for "+site+": ", ip.String()+mask)
					}
				} else {
					subnetsList = append(subnetsList, ip.String()+mask)
					fmt.Println("IP addresses for "+site+": ", ip.String()+mask)
				}
			}
		}
	}
	return subnetsList
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func runUpdates(testMode bool) {
	resolvedSubnets := PerformDNSLookups(ActiveConfig.Sites)
	resolvedSubnets = append(resolvedSubnets, ActiveConfig.Subnets...)

	// we might have some overlap, so let's dedupe
	resolvedSubnets = unique(resolvedSubnets)

	subnetsString := strings.Join(resolvedSubnets, ",")
	fullCommand := fmt.Sprintf(ActiveConfig.TailscaleCommand, subnetsString)
	fmt.Println("Tailscale command: ", fullCommand)
	if !testMode {
		commandTokens := strings.Split(fullCommand, " ")
		cmd := exec.Command(commandTokens[0], commandTokens[1:]...)
		output, err := cmd.Output()
		fmt.Println(string(output))

		if err != nil {
			fmt.Printf("Failed to run command: %v\n", err)
			if ActiveConfig.Slack.Enabled {
				slack.WebhookURL = ActiveConfig.Slack.WebhookURL
				slack.PostError(err)
			}
			os.Exit(1)
		}
		fmt.Println(string(output))
		fmt.Println("Trying to update Approved Subnets...")
		setTailscaleApprovedSubnets(resolvedSubnets)
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostRouteUpdate(resolvedSubnets, ActiveConfig.TailscaleclientId)
		}

	} else {
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostRouteUpdate(resolvedSubnets, ActiveConfig.TailscaleclientId)
		}
		fmt.Println("In test mode, not running command")

	}
}

func getTailsScaleClientRouteSettings() {
	urlTemplate := "https://api.tailscale.com/api/v2/device/%s/routes"
	url := fmt.Sprintf(urlTemplate, ActiveConfig.TailscaleclientId)
	fmt.Println("URL:>", url)
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}
		return
	}

	req.Header.Set("Authorization", "Bearer "+ActiveConfig.TailscaleKey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}
		return
	}

	rawMessage := json.RawMessage(body)

	// Marshal the raw message with indentation
	prettyJSON, err := json.MarshalIndent(rawMessage, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}
		return
	}

	// Print the pretty JSON
	fmt.Println(string(prettyJSON))
}

func setTailscaleApprovedSubnets(subnets []string) {
	urlTemplate := "https://api.tailscale.com/api/v2/device/%s/routes"
	url := fmt.Sprintf(urlTemplate, ActiveConfig.TailscaleclientId)
	fmt.Println("URL:>", url)

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
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ActiveConfig.TailscaleKey)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		fmt.Println("Response:", string(body))
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}

		return
	}

}

func main() {

	rootCmd := &cobra.Command{
		Use: "tailscale-route-tiler",
		Long: `This is a helper tool to getnerate a list of subnets for tailscale and the run the tailscale 
		command to update the routes.  
		`,
		Short: "A tailscale helper tool for subnets",
		Run: func(cmd *cobra.Command, args []string) {
			// This is the default action when no command is provided
			cmd.Help()
		},
	}

	// Flags
	var ConfigFile string
	rootCmd.PersistentFlags().StringVarP(&ConfigFile, "config", "c", "", "Specify the configuration file")

	// version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of tailscale-route-tiler",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("tailscale-route-tiler %s, commit %s, built at %s\n", version, commit, date)
		},
	})

	// Run Command
	var testMode bool

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the tailscale command to update the routes",
		Run: func(cmd *cobra.Command, args []string) {
			ReadYAML(ConfigFile)
			runUpdates(testMode)
		},
	}

	runCmd.Flags().BoolVarP(&testMode, "test", "t", false, "Run in test mode")
	rootCmd.AddCommand(runCmd)

	// Get Client Routes Command
	getClientRoutes := &cobra.Command{
		Use:   "get-client-routes",
		Short: "Get the current routes for the client",
		Run: func(cmd *cobra.Command, args []string) {
			ReadYAML(ConfigFile)
			getTailsScaleClientRouteSettings()
		},
	}
	rootCmd.AddCommand(getClientRoutes)

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
