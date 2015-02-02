package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// SelectServiceURL gets a eureka instance based on the connection's load
// balancing scheme.
// TODO: Make this not just pick a random one.
func (e *EurekaConnection) SelectServiceURL() string {
	if e.discoveryTtl == nil {
		e.discoveryTtl = make(chan struct{}, 1)
	}
	if e.DNSDiscovery && len(e.discoveryTtl) == 0 {
		servers, ttl, err := discoverDNS(e.DiscoveryZone, e.ServicePort)
		if err != nil {
			return choice(e.ServiceUrls)
		}
		e.discoveryTtl <- struct{}{}
		time.AfterFunc(ttl, func() {
			// At the end of the timeout, empty the channel so that the next
			// SelectServiceURL call will refresh the DNS info
			<-e.discoveryTtl
		})
		e.ServiceUrls = servers
	}
	return choice(e.ServiceUrls)
}

func choice(options []string) string {
	if len(options) == 0 {
		log.Fatal("There are no ServiceUrls to choose from, bailing out")
	}
	return options[rand.Int()%len(options)]
}

// NewConnFromConfigFile sets up a connection object based on a config in
// specified path
func NewConnFromConfigFile(location string) (c EurekaConnection, err error) {
	cfg, err := ReadConfig(location)
	if err != nil {
		log.Error("Problem reading config %s error: %s", location, err.Error())
		return c, err
	}
	return NewConnFromConfig(cfg), nil
}

// NewConnFromConfig will, given a Config struct, return a connection based on
// those options
func NewConnFromConfig(conf Config) (c EurekaConnection) {
	c.ServiceUrls = conf.Eureka.ServiceUrls
	c.ServicePort = conf.Eureka.ServerPort
	if len(c.ServiceUrls) == 0 && len(conf.Eureka.ServerDNSName) > 0 {
		c.ServiceUrls = []string{conf.Eureka.ServerDNSName}
	}
	c.Timeout = time.Duration(conf.Eureka.ConnectTimeoutSeconds) * time.Second
	c.PollInterval = time.Duration(conf.Eureka.PollIntervalSeconds) * time.Second
	c.PreferSameZone = conf.Eureka.PreferSameZone
	if conf.Eureka.UseDNSForServiceUrls {
		log.Warning("UseDNSForServiceUrls is an experimental option")
		c.DNSDiscovery = true
		c.DiscoveryZone = conf.Eureka.DNSDiscoveryZone
	}
	return c
}

// NewConn is a default connection with just a list of ServiceUrls. Most basic
// way to make a new connection. Generally only if you know what you're doing
// and are going to do the configuration yourself some other way.
func NewConn(address ...string) (e EurekaConnection) {
	e.ServiceUrls = address
	return e
}

// UpdateApp creates a goroutine that continues to keep an application updated
// with its status in Eureka
func (e *EurekaConnection) UpdateApp(app *Application) {
	go func() {
		for {
			log.Notice("Updating app %s", app.Name)
			err := e.readAppInto(app)
			if err != nil {
				log.Error("Failure updating %s in goroutine", app.Name)
			}
			<-time.After(time.Duration(e.PollInterval) * time.Second)
		}
	}()
}
