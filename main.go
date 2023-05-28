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

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var subnets []string

// Config is a struct for our YAML data
type Config struct {
	Subnets           []string `yaml:"subnets"`
	Sites             []string `yaml:"sites"`
	TailscaleCommand  string   `yaml:"TailscaleCommand"`
	EnableIpv6        bool     `yaml:"EnableIpv6"`
	TailscaleclientId string   `yaml:"TailscaleclientId"`
	TailscaleKey      string   `yaml:"TailscaleKey"`
}

// ReadYAML reads the YAML configuration file
func ReadYAML(filename string) (*Config, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func PerformDNSLookups(sites []string, config *Config) []string {
	var subnetsList []string
	for _, site := range sites {
		ips, err := net.LookupIP(site)
		if err != nil {
			fmt.Println("Error: ", site, err.Error())
		} else {
			for _, ip := range ips {
				mask := "/32"
				if ip.To4() == nil { // it's IPv6
					mask = "/128"
					if config.EnableIpv6 {
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

func runUpdates(config *Config, testMode bool) {
	resolvedSubnets := PerformDNSLookups(config.Sites, config)
	resolvedSubnets = append(resolvedSubnets, config.Subnets...)
	subnetsString := strings.Join(resolvedSubnets, ",")
	fullCommand := fmt.Sprintf(config.TailscaleCommand, subnetsString)
	fmt.Println("Tailscale command: ", fullCommand)
	if !testMode {
		commandTokens := strings.Split(fullCommand, " ")
		cmd := exec.Command(commandTokens[0], commandTokens[1:]...)
		output, err := cmd.Output()
		fmt.Println(string(output))

		if err != nil {
			fmt.Printf("Failed to run command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
		fmt.Println("Trying to update Approved Subnets...")
		setTailscaleApprovedSubnets(config, resolvedSubnets)

	} else {
		fmt.Println("In test mode, not running command")
	}
}

func getTailsScaleClientRouteSettings(config *Config) {
	urlTemplate := "https://api.tailscale.com/api/v2/device/%s/routes"
	url := fmt.Sprintf(urlTemplate, config.TailscaleclientId)
	// fmt.Println("Client ID: ", config)
	fmt.Println("URL:>", url)
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+config.TailscaleKey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	rawMessage := json.RawMessage(body)

	// Marshal the raw message with indentation
	prettyJSON, err := json.MarshalIndent(rawMessage, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// Print the pretty JSON
	fmt.Println(string(prettyJSON))
}

func setTailscaleApprovedSubnets(config *Config, subnets []string) {
	urlTemplate := "https://api.tailscale.com/api/v2/device/%s/routes"
	url := fmt.Sprintf(urlTemplate, config.TailscaleclientId)
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
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.TailscaleKey)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		fmt.Println("Response:", string(body))

		return
	}

}

func main() {
	rootCmd := &cobra.Command{
		Use:   "tailscale-route-tiler",
		Long:  "This is a helper tool to getnerate a list of subnets for tailscale and the run the tailscale command to update the routes",
		Short: "A tailscale helper tool for subnets",
		Run: func(cmd *cobra.Command, args []string) {
			// This is the default action when no command is provided
			cmd.Help()
		},
	}

	// Flags
	var configFile string
	var testMode bool

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the tailscale command to update the routes",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := ReadYAML(configFile)
			if err != nil {
				fmt.Println("Error reading YAML file: ", err)
				os.Exit(1)
			}
			runUpdates(config, testMode)
		},
	}

	runCmd.Flags().StringVarP(&configFile, "config", "c", "", "Specify the configuration file")
	runCmd.Flags().BoolVarP(&testMode, "test", "t", false, "Run in test mode")

	getClientRoutes := &cobra.Command{
		Use:   "get-client-routes",
		Short: "Get the current routes for the client",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := ReadYAML(configFile)
			if err != nil {
				fmt.Println("Error reading YAML file: ", err)
				os.Exit(1)
			}
			getTailsScaleClientRouteSettings(config)
		},
	}

	getClientRoutes.Flags().StringVarP(&configFile, "config", "c", "", "Specify the configuration file")

	// Add commands to the root command
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(getClientRoutes)

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
