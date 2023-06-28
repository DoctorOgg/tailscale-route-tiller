package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"tailscale-route-tiller/config"
	"tailscale-route-tiller/slack"
	"tailscale-route-tiller/tailscale"
	"tailscale-route-tiller/utils"
	"tailscale-route-tiller/worker"

	"github.com/spf13/cobra"
)

var subnets []string

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func runUpdates(testMode bool, config config.Config) {

	resolvedSubnets, _, err := utils.PerformDNSLookups(config.Sites, config.EnableIpv6)
	if err != nil {
		log.Println("Error: ", err.Error())
		slack.PostError(err)
		os.Exit(1)
	}

	resolvedSubnets = append(resolvedSubnets, config.Subnets...)

	// we might have some overlap, so let's dedupe
	resolvedSubnets = utils.Unique(resolvedSubnets)

	subnetsString := strings.Join(resolvedSubnets, ",")

	fullCommand := fmt.Sprintf(config.TailscaleCommand, subnetsString)

	if !testMode {
		output := utils.RunShellCommand(fullCommand, testMode)
		log.Println(output)
		log.Println("Trying to update Approved Subnets...")

		err = tailscale.SetTailscaleApprovedSubnets(resolvedSubnets)
		if err != nil {
			log.Println("Error: ", err.Error())
			slack.PostError(err)
			os.Exit(1)
		}

		slack.PostRouteUpdate(resolvedSubnets, config.TailscaleclientId)

	} else {
		slack.PostRouteUpdate(resolvedSubnets, config.TailscaleclientId)
		log.Println("In test mode, not running command")

	}
}

func runGetTailsScaleClientRouteSettings(config config.Config) {

	output, err := tailscale.GetTailsScaleClientRouteSettings()
	if err != nil {
		log.Println("Error: ", err.Error())
		slack.PostError(err)
		os.Exit(1)
	}

	log.Println(string(output))
}

func initConfig(configFile string) {
	config.ReadYAML(configFile)
	slack.WebhookURL = config.ActiveConfig.Slack.WebhookURL
	slack.Enabled = config.ActiveConfig.Slack.Enabled
	tailscale.TailscaleKey = config.ActiveConfig.TailscaleKey
	tailscale.TailScaleClientId = config.ActiveConfig.TailscaleclientId
}

func main() {
	log.Println("Starting tailscale-route-tiler", "version", version, "commit", commit, "date", date)

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
			initConfig(ConfigFile)
			runUpdates(testMode, *config.ActiveConfig)
		},
	}

	runCmd.Flags().BoolVarP(&testMode, "test", "t", false, "Run in test mode")
	rootCmd.AddCommand(runCmd)

	// worker Command
	workerCmd := &cobra.Command{
		Use:   "worker",
		Short: "Run in worker mode, will run periodically, based on the lowest record TTL",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig(ConfigFile)
			worker.Run(testMode, *config.ActiveConfig)
		},
	}

	workerCmd.Flags().BoolVarP(&testMode, "test", "t", false, "Run in test mode")
	rootCmd.AddCommand(workerCmd)

	// Get Client Routes Command
	getClientRoutes := &cobra.Command{
		Use:   "get-client-routes",
		Short: "Get the current routes for the client",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig(ConfigFile)
			runGetTailsScaleClientRouteSettings(*config.ActiveConfig)
		},
	}
	rootCmd.AddCommand(getClientRoutes)

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
