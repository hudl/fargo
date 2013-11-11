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
	"github.com/ryansb/fargo"
	. "launchpad.net/gocheck"
)

func (s *S) TestGetAllApps(c *C) {
	e := fargo.NewConn("http", "127.0.0.1", "8080")
	a, _ := e.GetApps()
	c.Assert(a["EUREKA"].Instances[0].HostName, Equals, "localhost.localdomain")
	c.Assert(a["EUREKA"].Instances[0].IpAddr, Equals, "127.0.0.1")
}

func (s *S) TestGetAppInstances(c *C) {
	e := fargo.NewConn("http", "127.0.0.1", "8080")
	a, _ := e.GetApp("EUREKA")
	c.Assert(a.Instances[0].HostName, Equals, "localhost.localdomain")
	c.Assert(a.Instances[0].IpAddr, Equals, "127.0.0.1")
}

func (s *S) TestRegisterFakeInstance(c *C) {
	e := fargo.NewConn("http", "127.0.0.1", "8080")
	i := fargo.Instance{
		HostName:         "i-123456",
		Port:             9090,
		App:              "TESTAPP",
		IpAddr:           "127.0.0.10",
		VipAddress:       "127.0.0.10",
		DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
		SecureVipAddress: "127.0.0.10",
		Status:           fargo.UP,
	}
	err := e.RegisterInstance(&i)
	c.Assert(err, IsNil)
}

func (s *S) TestCheckin(c *C) {
	e := fargo.NewConn("http", "127.0.0.1", "8080")
	i := fargo.Instance{
		HostName:         "i-123456",
		Port:             9090,
		App:              "TESTAPP",
		IpAddr:           "127.0.0.10",
		VipAddress:       "127.0.0.10",
		DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
		SecureVipAddress: "127.0.0.10",
		Status:           fargo.UP,
	}
	err := e.HeartBeatInstance(&i)
	c.Assert(err, IsNil)
}
