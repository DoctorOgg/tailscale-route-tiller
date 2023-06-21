package utils

import (
	"fmt"
	"net"
)

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

type IPSubnet struct {
	Network net.IPNet
}

func PerformDNSLookups(sites []string, enableIPv6 bool) ([]string, error) {

	var subnetsList []string

	for _, site := range sites {

		ips, err := net.LookupIP(site)

		if err != nil {
			fmt.Println("Error: ", site, err.Error())
			return nil, err
		}

		for _, ip := range ips {
			mask := "/32"
			if ip.To4() == nil { // it's IPv6
				if enableIPv6 {
					mask = "/128"
					subnetsList = append(subnetsList, ip.String()+mask)
				}
			} else {
				subnetsList = append(subnetsList, ip.String()+mask)
			}
		}

	}
	return subnetsList, nil

}
