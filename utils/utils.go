package utils

import (
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

// type IPSubnet struct {
// 	Network net.IPNet
// }

type IPWithTTL struct {
	IP  string
	TTL time.Duration
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

func lookupIPsWithTTL(host string) ([]IPWithTTL, error) {
	// Get system DNS server
	dnsServer, err := getSystemDNS()
	if err != nil {
		return nil, err
	}

	dnsClient := new(dns.Client)
	dnsMsg := new(dns.Msg)

	dnsMsg.SetQuestion(dns.Fqdn(host), dns.TypeANY)
	dnsMsg.RecursionDesired = true

	r, _, err := dnsClient.Exchange(dnsMsg, dnsServer+":53")
	if err != nil {
		return nil, err
	}

	if len(r.Answer) < 1 {
		return nil, fmt.Errorf("no answer for host")
	}

	var results []IPWithTTL
	for _, ans := range r.Answer {
		switch record := ans.(type) {
		case *dns.A:
			results = append(results, IPWithTTL{IP: record.A.String(), TTL: time.Duration(record.Hdr.Ttl) * time.Second})
		case *dns.AAAA:
			results = append(results, IPWithTTL{IP: record.AAAA.String(), TTL: time.Duration(record.Hdr.Ttl) * time.Second})
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no A or AAAA record found for host")
	}

	return results, nil
}

func PerformDNSLookups(sites []string, enableIPv6 bool) ([]string, uint32, error) {

	var subnetsList []string
	var lowestTTL uint32 = 60
	var ipv4Mask string = "/32"
	var ipv6Mask string = "/128"

	for _, site := range sites {

		results, err := lookupIPsWithTTL(site)
		if err != nil {
			fmt.Println("Error: ", site, err.Error())
			return nil, lowestTTL, err
		}

		for _, result := range results {
			if result.TTL < time.Duration(lowestTTL)*time.Second {
				lowestTTL = uint32(result.TTL.Seconds())
			}

			// check if it's IPv6
			if net.ParseIP(result.IP).To4() == nil {
				if enableIPv6 {
					subnetsList = append(subnetsList, result.IP+ipv6Mask)
				}
			} else {
				subnetsList = append(subnetsList, result.IP+ipv4Mask)
			}

		}

	}

	if lowestTTL < 60 {
		lowestTTL = 60
	}
	return subnetsList, lowestTTL, nil

}
