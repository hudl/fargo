package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"fmt"
	"github.com/franela/goreq"
	"github.com/miekg/dns"
	"time"
)

const azURL = "http://169.254.169.254/latest/meta-data/placement/availability-zone"

var ErrNotInAWS = fmt.Errorf("Not in AWS")

func discoverDNS(domain string) (servers []string, ttl time.Duration, err error) {
	r, _ := region()

	// all DNS queries must use the FQDN
	domain = "txt." + r + "." + dns.Fqdn(domain)
	if _, ok := dns.IsDomainName(domain); !ok {
		err = fmt.Errorf("invalid domain name: '%s' is not a domain name", domain)
		return
	}
	regionRecords, ttl, err := findTXT(domain)
	if err != nil {
		return
	}

	for _, az := range regionRecords {
		s, _, er := findTXT("txt." + dns.Fqdn(az))
		if er != nil {
			continue
		}
		servers = append(servers, s...)
	}
	return
}

func findTXT(fqdn string) ([]string, time.Duration, error) {
	defaultTTL := 120 * time.Second
	query := new(dns.Msg)
	query.SetQuestion(fqdn, dns.TypeTXT)
	response, err := dns.Exchange(query, dnsServerAddr)
	if err != nil {
		log.Error("Failure resolving name %s err=%s", fqdn, err.Error())
		return nil, defaultTTL, err
	}
	if len(response.Answer) < 1 {
		err := fmt.Errorf("no Eureka discovery TXT record returned for name=%s", fqdn)
		log.Error("no answer for name=%s err=%s", fqdn, err.Error())
		return nil, defaultTTL, err
	}
	if response.Answer[0].Header().Rrtype != dns.TypeTXT {
		err := fmt.Errorf("did not receive TXT record back from query specifying TXT record. This should never happen.")
		log.Error("Failure resolving name %s err=%s", fqdn, err.Error())
		return nil, defaultTTL, err
	}
	txt := response.Answer[0].(*dns.TXT)
	ttl := response.Answer[0].Header().Ttl

	return txt.Txt, time.Duration(ttl) * time.Second, nil
}

var dnsServerAddr string

func init() {
	// Find a DNS server using the OS resolv.conf
	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	dnsServerAddr = config.Servers[0] + ":" + config.Port
}

func region() (string, error) {
	zone, err := availabilityZone()
	if err != nil {
		log.Error("Could not retrieve availability zone err=%s", err.Error())
		return "us-east-1", err
	}
	return zone[:len(zone)-1], nil
}

// defaults to us-east-1 if there's a problem
func availabilityZone() (string, error) {
	response, err := goreq.Request{Uri: azURL}.Do()
	if err != nil {
		return "", err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := response.Body.ToString()
		return "", fmt.Errorf("bad response code: code %d does not indicate successful request, body=%s",
			response.StatusCode,
			body,
		)
	}
	zone, err := response.Body.ToString()
	if err != nil {
		return "", err
	}
	return zone[:len(zone)-1], nil
}
