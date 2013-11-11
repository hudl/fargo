package fargo

/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2013 Ryan S. Brown <sb@ryansb.com>
 * Copyright (c) 2013 Hudl <@Hudl>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to
 * deal in the Software without restriction, including without limitation the
 * rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
 * sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
 * IN THE SOFTWARE.
 */

import (
	"code.google.com/p/gcfg"
)

type Config struct {
	AWS    aws
	Eureka eureka
}

type aws struct {
	AccessKeyId     string
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
	ConnectTimeoutSeconds int32    // default 10s
	UseDnsForServiceUrls  bool     // default false
	ServerDnsName         string   // default ""
	ServiceUrls           []string // default []
	ServerPort            int32    // default 7001
	PollIntervalSeconds   int32    // default 30
	EnableDelta           bool     // TODO: Support querying for deltas
	PreferSameZone        bool     // default false
	RegisterWithEureka    bool     // default false
}

func ReadConfig(loc string) (conf Config, err error) {
	err = gcfg.ReadFileInto(&conf, loc)
	if err != nil {
		log.Critical("Unable to read config file Error: %s", err.Error())
		return conf, err
	}
	conf.FillDefaults()
	return conf, nil
}

func (c *Config) FillDefaults() {
	// TODO: Read in current Availability Zone if in AWS (DC==Amazon)
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
