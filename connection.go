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
	"math/rand"
)

func (e *EurekaConnection) SelectServiceUrl() string {
	return e.ServiceUrls[rand.Int31n(int32(len(e.ServiceUrls)))]
}

func NewConnFromConfigFile(location string) (c EurekaConnection, err error) {
	cfg, err := ReadConfig(location)
	if err != nil {
		log.Error("Problem reading config %s error: %s", location, err.Error())
		return c, nil
	}
	return NewConnFromConfig(cfg), nil
}
func NewConnFromConfig(conf Config) (c EurekaConnection) {
	if conf.Eureka.UseDnsForServiceUrls {
		//TODO: Read service urls from DNS TXT records
		log.Critical("ERROR: UseDnsForServiceUrls option unsupported.")
	}
	c.ServiceUrls = conf.Eureka.ServiceUrls
	if len(c.ServiceUrls) == 0 && len(conf.Eureka.ServerDnsName) > 0 {
		c.ServiceUrls = []string{conf.Eureka.ServerDnsName}
	}
	c.Timeout = conf.Eureka.ConnectTimeoutSeconds
	c.PollInterval = conf.Eureka.PollIntervalSeconds
	c.PreferSameZone = conf.Eureka.PreferSameZone
	return c
}

func NewConn(address ...string) (c EurekaConnection) {
	c.ServiceUrls = address
	return c
}
