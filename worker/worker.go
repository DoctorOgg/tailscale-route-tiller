package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"tailscale-route-tiller/config"
	"tailscale-route-tiller/slack"
	"tailscale-route-tiller/tailscale"
	"tailscale-route-tiller/utils"
	"time"

	"tailscale-route-tiller/cloudwatchevent"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var TestMode bool = false
var Command string

func parseCloudWatchEvent(message *sqs.Message) (*cloudwatchevent.CloudTrailEvent, error) {
	var event cloudwatchevent.CloudTrailEvent
	err := json.Unmarshal([]byte(*message.Body), &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func Run(testMode bool, config config.Config) {

	// Initialize a session in us-west-2 region that the SDK will use to load credentials
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)

	if err != nil {
		log.Fatalf("failed to create session, %v", err)
	}

	// Create a SQS service client
	svc := sqs.New(sess)

	for {
		// Receive a message from the SQS queue
		result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            &config.SQS.QueueURL,
			MaxNumberOfMessages: aws.Int64(1),
			WaitTimeSeconds:     aws.Int64(20),
		})

		if err != nil {
			log.Fatalf("Unable to receive message from queue %q, %v.", config.SQS.QueueURL, err)
		}

		if len(result.Messages) > 0 {
			// Call a function to process the message here
			// processMessage(result.Messages[0])

			message := result.Messages[0]

			if testMode {
				log.Println("Test mode enabled. Message: ", *message.Body)
			}

			// lets parse the message and hand off to the runUpdates function
			event, err := parseCloudWatchEvent(message)
			if err != nil {
				log.Printf("Error parsing CloudWatch event: %v", err)
				continue // Skip this message or handle the error as appropriate
			}

			// before running the udpate, we should wait for dns to settle

			// lets wait for 2 minutes for DNS to settle
			log.Println("Waiting for DNS to settle...")
			time.Sleep(2 * time.Minute)

			runUpdates(testMode, config, event)

			// Delete the message from the queue after processing
			_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      &config.SQS.QueueURL,
				ReceiptHandle: message.ReceiptHandle,
			})

			if err != nil {
				log.Fatalf("Failed to delete message from queue, %v", err)
			}
		}
	}
}

func runUpdates(testMode bool, config config.Config, event *cloudwatchevent.CloudTrailEvent) {

	resolvedSubnets, _, err := utils.PerformDNSLookupsWithTTL(config.Sites, config.EnableIpv6)
	if err != nil {
		log.Println("Error: ", err.Error())
		slack.PostError(err)
	}

	// Get the final list of subnets to approve
	resolvedSubnets = append(resolvedSubnets, config.Subnets...)
	resolvedSubnets = utils.Unique(resolvedSubnets)

	log.Println("Resolved subnets: ", resolvedSubnets)

	// format subnets as a string for the tailscale command
	subnetsString := strings.Join(resolvedSubnets, ",")

	// combine the command with the subnets
	fullCommand := fmt.Sprintf(config.TailscaleCommand, subnetsString)

	if testMode {
		log.Println("Test mode enabled. Command: ", fullCommand)
	} else {
		// run the command
		output := utils.RunShellCommand(fullCommand, testMode)

		// log the output
		log.Println(string(output))
	}

	networkDescription := event.Detail.RequestParameters.Description
	slack.PostRouteUpdateSQS(networkDescription, config.TailscaleclientId)

	if testMode {
		log.Println("Test mode enabled, not updating tailscale routes.")
	} else {
		err = tailscale.SetTailscaleApprovedSubnets(resolvedSubnets)
		if err != nil {
			log.Println("Error: ", err.Error())
			slack.PostError(err)
			os.Exit(1)
		}
	}
}
