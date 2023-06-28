package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/miekg/dns"
)

type IPWithTTL struct {
	IP  string
	TTL int
}

func Unique(slice []string) []string {
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

func getSystemDNS() (string, error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil || len(config.Servers) == 0 {
		return "", fmt.Errorf("Could not find system DNS: %v", err)

	}
	return config.Servers[0], nil
}

func lookupIPsWithTTL(host string, enableIpv6 bool) ([]IPWithTTL, error) {
	var results []IPWithTTL

	dnsServer, err := getSystemDNS()
	if err != nil {
		return nil, err
	}
	dnsClient := new(dns.Client)
	dnsMsg := new(dns.Msg)

	if enableIpv6 {
		dnsMsg.SetQuestion(dns.Fqdn(host), dns.TypeAAAA)
	} else {
		dnsMsg.SetQuestion(dns.Fqdn(host), dns.TypeA)
	}
	dnsMsg.RecursionDesired = true

	r, _, err := dnsClient.Exchange(dnsMsg, dnsServer+":53")
	if err != nil {
		return nil, err
	}

	for _, ans := range r.Answer {
		switch record := ans.(type) {
		case *dns.A:
			results = append(results, IPWithTTL{IP: record.A.String(), TTL: int(record.Hdr.Ttl)})
		case *dns.AAAA:
			results = append(results, IPWithTTL{IP: record.AAAA.String(), TTL: int(record.Hdr.Ttl)})
		}
	}

	return results, nil
}

func PerformDNSLookups(sites []string, enableIPv6 bool) ([]string, int, error) {

	var subnetsList []string
	var lowestTTL int = 60
	var ipv4Mask string = "/32"
	var ipv6Mask string = "/128"

	for _, site := range sites {

		results, err := lookupIPsWithTTL(site, false)
		if err != nil {
			log.Println("Error: ", site, err.Error())
			return nil, lowestTTL, err
		}

		if len(results) > 0 {
			for _, result := range results {
				if result.TTL < lowestTTL {
					lowestTTL = result.TTL
				}
				subnetsList = append(subnetsList, result.IP+ipv4Mask)
			}
		}
	}

	// now if we need ipv6 addresses
	if enableIPv6 {
		for _, site := range sites {

			results, err := lookupIPsWithTTL(site, true)
			if err != nil {
				log.Println("Error: ", site, err.Error())
				return nil, lowestTTL, err
			}

			if len(results) > 0 {
				for _, result := range results {
					if result.TTL < lowestTTL {
						lowestTTL = result.TTL
					}
					subnetsList = append(subnetsList, result.IP+ipv6Mask)
				}
			}

		}
	}

	if lowestTTL < 60 {
		lowestTTL = 60
	}

	return subnetsList, lowestTTL, nil

}

func RunShellCommand(command string, testMode bool) string {
	log.Println("command: ", command)
	if !testMode {
		commandTokens := strings.Split(command, " ")
		cmd := exec.Command(commandTokens[0], commandTokens[1:]...)
		output, err := cmd.Output()

		if err != nil {
			log.Printf("Failed to run command: %v\n", err)
			os.Exit(1)
		}
		return string(output)
	}
	return "No Output, Test Mode"
}
