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
	return e.ServiceUrls[rand.Int()%len(e.ServiceUrls)]
}

// NewConnFromConfigFile sets up a connection object based on a config in
// specified path
func NewConnFromConfigFile(location string) (c EurekaConnection, err error) {
	cfg, err := ReadConfig(location)
	if err != nil {
		log.Error("Problem reading config %s error: %s", location, err.Error())
		return c, nil
	}
	return NewConnFromConfig(cfg), nil
}

// NewConnFromConfig will, given a Config struct, return a connection based on
// those options
func NewConnFromConfig(conf Config) (c EurekaConnection) {
	if conf.Eureka.UseDNSForServiceUrls {
		//TODO: Read service urls from DNS TXT records
		log.Critical("ERROR: UseDNSForServiceUrls option unsupported.")
	}
	c.ServiceUrls = conf.Eureka.ServiceUrls
	if len(c.ServiceUrls) == 0 && len(conf.Eureka.ServerDNSName) > 0 {
		c.ServiceUrls = []string{conf.Eureka.ServerDNSName}
	}
	c.Timeout = time.Duration(conf.Eureka.ConnectTimeoutSeconds) * time.Second
	c.PollInterval = time.Duration(conf.Eureka.PollIntervalSeconds) * time.Second
	c.PreferSameZone = conf.Eureka.PreferSameZone
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
