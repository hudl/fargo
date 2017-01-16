package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import "github.com/hudl/fargo"

func makeConnection() fargo.EurekaConnection {
	var c fargo.Config
	c.Eureka.ServiceUrls = []string{"http://172.17.0.2:8080/eureka/v2"}
	c.Eureka.ConnectTimeoutSeconds = 10
	c.Eureka.PollIntervalSeconds = 30
	c.Eureka.Retries = 3
	return fargo.NewConnFromConfig(c)
}
