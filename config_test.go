package fargo_test

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
	"github.com/hudl/fargo"
	. "launchpad.net/gocheck"
)

func (s *S) TestConfigDefaults(c *C) {
	conf, err := fargo.ReadConfig("./config_sample/blank.gcfg")
	c.Check(err, IsNil)
	c.Check(conf.Eureka.InTheCloud, Equals, false)
	c.Check(conf.Eureka.ConnectTimeoutSeconds, Equals, int32(10))
	c.Check(conf.Eureka.UseDnsForServiceUrls, Equals, false)
	c.Check(conf.Eureka.ServerDnsName, Equals, "")
	c.Check(len(conf.Eureka.ServiceUrls), Equals, 0)
	c.Check(conf.Eureka.ServerPort, Equals, int32(7001))
	c.Check(conf.Eureka.PollIntervalSeconds, Equals, int32(30))
	c.Check(conf.Eureka.EnableDelta, Equals, false)
	c.Check(conf.Eureka.PreferSameZone, Equals, false)
	c.Check(conf.Eureka.RegisterWithEureka, Equals, false)
}

func (s *S) TestLocalConfig(c *C) {
	conf, err := fargo.ReadConfig("./config_sample/local.gcfg")
	c.Check(err, IsNil)
	c.Check(conf.Eureka.InTheCloud, Equals, false)
	c.Check(conf.Eureka.ConnectTimeoutSeconds, Equals, int32(2))
	c.Check(conf.Eureka.ServiceUrls, DeepEquals, []string{"things", "stuff"})
	c.Check(conf.Eureka.UseDnsForServiceUrls, Equals, false)
}
