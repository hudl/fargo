package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"gopkg.in/gcfg.v1"
)

// Config is a base struct to be read by code.google.com/p/gcfg
type Config struct {
	AWS    aws
	Eureka eureka
}

type aws struct {
	AccessKeyID     string
	SecretAccessKey string
	// zones if running in AWS specifies as [us-east-1a, us-east-1b]
	AvailabilityZones []string
	// service urls for individual zones, ex [eureka1.east1a.my.com, eureka2.east1a.my.com]
	ServiceUrlsEast1a []string
	ServiceUrlsEast1b []string
	ServiceUrlsEast1c []string
	ServiceUrlsEast1d []string
	ServiceUrlsEast1e []string
	Region            string // unused. Currently only set up for us-east-1
}

type eureka struct {
	InTheCloud            bool     // default false
	ConnectTimeoutSeconds int      // default 10s
	UseDNSForServiceUrls  bool     // default false
	DNSDiscoveryZone      string   // default ""
	ServerDNSName         string   // default ""
	ServiceUrls           []string // default []
	ServerPort            int      // default 7001
	PollIntervalSeconds   int      // default 30
	EnableDelta           bool     // TODO: Support querying for deltas
	PreferSameZone        bool     // default false
	RegisterWithEureka    bool     // default false
	Retries               int      // default 3
}

// ReadConfig from a file location. Minimal error handling. Just bails and passes up
// an error if the file isn't found
func ReadConfig(loc string) (conf Config, err error) {
	err = gcfg.ReadFileInto(&conf, loc)
	if err != nil {
		log.Criticalf("Unable to read config file Error: %s", err.Error())
		return conf, err
	}
	conf.fillDefaults()
	return conf, nil
}

func (c *Config) fillDefaults() {
	// TODO: Read in current Availability Zone if in AWS (DC==Amazon)
	if c.Eureka.Retries == 0 {
		c.Eureka.Retries = 3
	}
	if c.Eureka.ConnectTimeoutSeconds == 0 {
		c.Eureka.ConnectTimeoutSeconds = 10
	}
	if c.Eureka.ServerPort == 0 {
		c.Eureka.ServerPort = 7001
	}
	if c.Eureka.PollIntervalSeconds == 0 {
		c.Eureka.PollIntervalSeconds = 30
	}
}
