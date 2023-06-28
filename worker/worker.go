package worker

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"tailscale-route-tiller/config"
	"tailscale-route-tiller/slack"
	"tailscale-route-tiller/tailscale"
	"tailscale-route-tiller/utils"
	"time"
)

type IPWithTTL struct {
	IP  string
	TTL time.Duration
}

var interval time.Duration = 1 * time.Second
var currentSubnets []string
var firstRun bool = true
var TestMode bool = false
var Command string

func getAddedElements(current, newArray []string) []string {
	// Create a map to store the elements of the current array
	currentMap := make(map[string]bool)
	for _, item := range current {
		currentMap[item] = true
	}

	// Iterate over the new array and check for elements not present in the current array
	var added []string
	for _, item := range newArray {
		if !currentMap[item] {
			added = append(added, item)
		}
	}

	return added
}

func getRemovedElements(current, newArray []string) []string {
	// Use the getAddedElements function by swapping the arrays
	return getAddedElements(newArray, current)
}

func Run(testMode bool, config config.Config) {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create a channel to control the main loop
	done := make(chan bool)

	// Start the goroutine
	go runUpdates(done, testMode, config)

	// Wait for termination signals
	<-sigChan

	// Signal the main loop to stop
	done <- true

	log.Println("Program terminated")

}

func runUpdates(done chan bool, testMode bool, config config.Config) {
	sleepChan := time.After(interval) // Start initial sleep duration

	for {
		select {
		case <-done:
			done <- true // Signal completion to the main function
			return
		case <-sleepChan:
			// Perform updates
			resolvedSubnets, lowestTTL, err := utils.PerformDNSLookups(config.Sites, config.EnableIpv6)
			if err != nil {
				log.Println("Error: ", err.Error())
				slack.PostError(err)
			}

			// Set the Interval to the lowest TTL unless lower than 60
			if lowestTTL > 120 {
				interval = time.Duration(lowestTTL) * time.Second
			} else {
				interval = 120 * time.Second
			}
			// fmt.Println("New Interval: ", interval)

			// Get the final list of subnets to approve
			resolvedSubnets = append(resolvedSubnets, config.Subnets...)
			resolvedSubnets = utils.Unique(resolvedSubnets)

			// Cases to handle: first run, no changes, changes
			if firstRun {
				currentSubnets = resolvedSubnets
				firstRun = false

				subnetsString := strings.Join(currentSubnets, ",")

				fullCommand := fmt.Sprintf(config.TailscaleCommand, subnetsString)
				output := utils.RunShellCommand(fullCommand, testMode)
				log.Println(string(output))
				err = tailscale.SetTailscaleApprovedSubnets(resolvedSubnets)
				if err != nil {
					log.Println("Error: ", err.Error())
					slack.PostError(err)
					os.Exit(1)
				}

				slack.PostRouteUpdate(resolvedSubnets, config.TailscaleclientId)
			} else if len(resolvedSubnets) == len(currentSubnets) {
				log.Println("No changes detected, moving along, New Interval: " + interval.String())
			} else {
				log.Println("Changes detected, updating, New Interval: " + interval.String())
				subnetsString := strings.Join(currentSubnets, ",")
				fullCommand := fmt.Sprintf(config.TailscaleCommand, subnetsString)
				output := utils.RunShellCommand(fullCommand, testMode)
				fmt.Println(string(output))
				err = tailscale.SetTailscaleApprovedSubnets(resolvedSubnets)
				if err != nil {
					log.Println("Error: ", err.Error())
					slack.PostError(err)
					os.Exit(1)
				}

				added := getAddedElements(currentSubnets, resolvedSubnets)
				removed := getRemovedElements(currentSubnets, resolvedSubnets)
				currentSubnets = resolvedSubnets
				slack.PostDiffUpdate(added, removed, config.TailscaleclientId)
			}

			sleepChan = time.After(interval) // Reset sleep duration
		}
	}
}
