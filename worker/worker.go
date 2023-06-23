package worker

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)



type IPWithTTL struct {
	IP  string
	TTL time.Duration
}

Interval := 60 * time.Second
currentSubnets := []string{}

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


func run() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	// Create a channel to control the main loop
	done := make(chan bool)

	// Start the goroutine outside the main function
	go runUpdates(done)

	// Wait for termination signals
	<-sigChan

	// Signal the main loop to stop
	done <- true

	fmt.Println("Program terminated")

}

func runUpdates(done <-chan bool) {
	for {
		select {
		case <-done:
			return
		default:
			// Ok lets do some work
			resolvedSubnets, lowestTTL, err := utils.PerformDNSLookups(config.Sites, config.EnableIpv6)
			if err != nil {
				fmt.Println("Error: ", err.Error())
				slack.PostError(err)
				os.Exit(1)
			}

			// set the Interval to the lowest TTL unless lower than 60
			if lowestTTL > 60 {
				Interval = lowestTTL * time.Second
			} else {
				Interval = 60 * time.Second
			}

		
			resolvedSubnets = append(resolvedSubnets, config.Subnets...)
			resolvedSubnets = utils.Unique(resolvedSubnets)


			added := getAddedElements(currentSubnets, resolvedSubnets)
			removed := getRemovedElements(currentSubnets, resolvedSubnets)

			currentSubnets = resolvedSubnets

			if len(added) > 0 && len(removed) > 0 {
				fmt.Println("Added: ", added)
				fmt.Println("Removed: ", removed)
				slack.PostRouteUpdate(resolvedSubnets, config.TailscaleclientId)
			
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
					slack.PostError(err)
					os.Exit(1)
				}

				fmt.Println(string(output))
				fmt.Println("Trying to update Approved Subnets...")

				err = tailscale.SetTailscaleApprovedSubnets(resolvedSubnets)
				if err != nil {
					fmt.Println("Error: ", err.Error())
					slack.PostError(err)
					os.Exit(1)
				}

				slack.PostRouteUpdate(resolvedSubnets, config.TailscaleclientId)

			} else {
				slack.PostRouteUpdate(resolvedSubnets, config.TailscaleclientId)
				fmt.Println("In test mode, not running command")
			}
		} else {
			fmt.Println("No changes detected, skipping tailscale command")
		}
			// Sleep for the lowest TTL unless less than 60 	
			time.Sleep(Interval)
		}
	}
}


